package hsm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/miekg/pkcs11"
	"github.com/spf13/viper"
)

var ErrInvalidParam = errors.New("invalid input parameter")
var ErrInvalidFilename = errors.New("invalid pkcs#11 module path/filename")
var ErrInitFailure = errors.New("pkcs#11 module init failure")
var ErrMissingSlot = errors.New("no slot defined")
var ErrSessionFailure = errors.New("unable to open session with slot 0")
var ErrLoginFailure = errors.New("unable to login to slot 0 with the specific pin")
var ErrQueryFailure = errors.New("unable to execute the requested PKCS#11 api call")

var CA_KEY_LABEL = "LITE Root CA"
var CA_KEY_UUID = []byte{0xd0, 0x61, 0x9e, 0x62, 0xdd, 0xa2, 0x43, 0xb4, 0xb5, 0x3c, 0x85, 0x0b, 0x07, 0xf0, 0x78, 0x1c}

// Protoype for pkcs#11 functions to run after connection validation in exec
type execFn func(p *pkcs11.Ctx, s pkcs11.SessionHandle) error

// exec performs basic error checking for PKCS#11 queries
func exec(fn execFn) error {
	module := viper.GetString("hsm.module")
	pin := viper.GetString("hsm.pin")

	if module == "" || pin == "" {
		return ErrInvalidParam
	}

	// Make sure the specified PKCS#11 shared module actually exists
	_, err := os.Stat(module)
	if err != nil {
		return ErrInvalidFilename
	}

	// Try to initialise the shared module
	p := pkcs11.New(module)
	if err = p.Initialize(); err != nil {
		return ErrInitFailure
	}

	defer p.Destroy()
	defer p.Finalize()

	// Make sure there is a slot defined
	// ToDo: Check for a specific label (LITEBoot)!
	slots, err := p.GetSlotList(true)
	if err != nil {
		return ErrMissingSlot
	}

	// Try to open a new session using slot 0
	session, err := p.OpenSession(slots[0], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		return ErrSessionFailure
	}
	defer p.CloseSession(session)

	// Attempt to login to slot 0 using the user context pin
	if err = p.Login(session, pkcs11.CKU_USER, pin); err != nil {
		return ErrLoginFailure
	}
	defer p.Logout(session)

	// PKCS#11 module seems to be OK, run the supplied function if present
	if fn != nil {
		err = fn(p, session)
	}

	return err
}

func TestConnection() error {
	// No exec function required since we just want to connect and exit
	return exec(nil)
}

func DisplayHSMInfo() error {
	var f = func(p *pkcs11.Ctx, s pkcs11.SessionHandle) error {
		info, err := p.GetInfo()
		if err != nil {
			return ErrQueryFailure
		}

		fmt.Printf("HSM:\n")
		fmt.Printf("  Name: %s\n", info.ManufacturerID)
		fmt.Printf("  Cryptoki: %v.%v\n", info.CryptokiVersion.Major, info.CryptokiVersion.Minor)
		fmt.Printf("  Library: %v.%v\n", info.LibraryVersion.Major, info.LibraryVersion.Minor)

		return nil
	}

	return exec(f)
}

func DisplaySlotInfo() error {
	var f = func(p *pkcs11.Ctx, s pkcs11.SessionHandle) error {
		sinfo, err := p.GetSessionInfo(s)
		if err != nil {
			return ErrQueryFailure
		}

		fmt.Printf("Slot 0:\n")
		fmt.Printf("  ID: 0x%x\n", sinfo.SlotID)

		tinfo, err := p.GetTokenInfo(sinfo.SlotID)
		if err != nil {
			return ErrQueryFailure
		}

		fmt.Printf("  Token label: %s\n", tinfo.Label)
		fmt.Printf("  Serial: %s\n", tinfo.SerialNumber)

		return nil
	}

	return exec(f)
}

func displayObject(p *pkcs11.Ctx, s pkcs11.SessionHandle, o pkcs11.ObjectHandle) error {
	attr, err := p.GetAttributeValue(s, o, []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, nil),
		pkcs11.NewAttribute(pkcs11.CKA_ID, nil),
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, nil), // CKK_RSA, CKK_EC, etc.
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, nil),    // CKO_PUBLIC_KEY, CKO_PRIVATE_KEY, etc.
	})
	if err != nil {
		return ErrQueryFailure
	}

	// Render the matching attributes
	for _, a := range attr {
		fmt.Printf("    ")
		switch a.Type {
		case pkcs11.CKA_KEY_TYPE:
			fmt.Printf("Key Type: ")
			switch a.Value[0] {
			case pkcs11.CKK_RSA:
				fmt.Printf("RSA\n")
			case pkcs11.CKK_EC:
				fmt.Printf("EC\n")
			default:
				fmt.Printf("%d\n", a.Value[0])
			}
		case pkcs11.CKA_LABEL:
			fmt.Printf("Label: %s\n", a.Value)
		case pkcs11.CKA_CLASS:
			fmt.Printf("Class: ")
			switch a.Value[0] {
			case pkcs11.CKO_PUBLIC_KEY:
				fmt.Printf("Public Key\n")
			case pkcs11.CKO_PRIVATE_KEY:
				fmt.Printf("Private Key\n")
			default:
				fmt.Printf("%d\n", a.Value[0])
			}
		case pkcs11.CKA_ID:
			fmt.Printf("ID: %x\n", a.Value)
		default:
			fmt.Printf("0x%04X: %s (len %d)\n", a.Type, hex.EncodeToString(a.Value), len(a.Value))
		}
	}

	return nil
}

func FindObjectsByLabel(label string, max int) error {
	var f = func(p *pkcs11.Ctx, s pkcs11.SessionHandle) error {
		// Set the search parameters
		t := []*pkcs11.Attribute{
			pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
		}

		if err := p.FindObjectsInit(s, t); err != nil {
			return ErrQueryFailure
		}

		// Get the first object that matches t
		o, _, err := p.FindObjects(s, max)
		if err != nil {
			return ErrQueryFailure
		}

		// Clean up
		if err = p.FindObjectsFinal(s); err != nil {
			return ErrQueryFailure
		}

		// Display matching object
		fmt.Printf("Found %d objects with label \"%s\"\n", len(o), label)
		for i, obj := range o {
			fmt.Printf("  Object %d:\n", i)
			if err := displayObject(p, s, obj); err != nil {
				return err
			}
		}

		return nil
	}

	return exec(f)
}

func FindObjectsByUUID(u uuid.UUID, max int) error {
	var f = func(p *pkcs11.Ctx, s pkcs11.SessionHandle) error {
		// Set the search parameters
		t := []*pkcs11.Attribute{
			pkcs11.NewAttribute(pkcs11.CKA_ID, []byte(u[0:])),
		}

		if err := p.FindObjectsInit(s, t); err != nil {
			return ErrQueryFailure
		}

		// Get the first object that matches t
		o, _, err := p.FindObjects(s, max)
		if err != nil {
			return ErrQueryFailure
		}

		// Clean up
		if err = p.FindObjectsFinal(s); err != nil {
			return ErrQueryFailure
		}

		// Display matching object
		fmt.Printf("Found %d objects with UUID \"%s\"\n", len(o), u.String())
		for i, obj := range o {
			fmt.Printf("  Object %d:\n", i)
			if err := displayObject(p, s, obj); err != nil {
				return err
			}
		}

		return nil
	}

	return exec(f)
}

func CreateRootCAKey() error {
	var f = func(p *pkcs11.Ctx, s pkcs11.SessionHandle) error {
		// Private key settins
		t := []*pkcs11.Attribute{
			pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
			pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_EC),
			pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
			pkcs11.NewAttribute(pkcs11.CKA_PRIVATE, true),
			pkcs11.NewAttribute(pkcs11.CKA_SENSITIVE, true),
			pkcs11.NewAttribute(pkcs11.CKA_LABEL, CA_KEY_LABEL),
			pkcs11.NewAttribute(pkcs11.CKA_ID, CA_KEY_UUID),
			pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
			pkcs11.NewAttribute(pkcs11.CKA_EXTRACTABLE, false),
			// TODO: Add other required attributes
		}

		_, err := p.CreateObject(s, t)
		if err != nil {
			return ErrQueryFailure
		}

		return nil
	}

	return exec(f)
}
