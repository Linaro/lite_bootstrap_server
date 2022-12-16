package hsm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/miekg/pkcs11"
	"github.com/spf13/viper"
)

var (
	// ErrNoHSM indicates that no HSM has been configured by the user.
	ErrNoHSM = errors.New("no HSM configured")

	// ErrInvalidFilename indicates that the pkcs#11 module
	// pathname is invalid.
	ErrInvalidFilename = errors.New("invalid pkcs#11 module path/filename")

	// ErrInitFailure indicates that we were unable to initialize
	// the pkcs#11 module.
	ErrInitFailure = errors.New("pkcs#11 module init failure")

	// ErrMissingSlot indicates that the HSM does not have the
	// appropriate slot for our keys.
	ErrMissingSlot = errors.New("no slot defined")

	// ErrSessionFailure indicates that we were unable to open a
	// session with the HSM.
	ErrSessionFailure = errors.New("unable to open session with slot 0")

	// ErrLoginFailure indicates that the login with the hsm
	// failed, likely because of an incorrect pin.
	ErrLoginFailure = errors.New("unable to login to slot 0 with the specific pin")

	// ErrQueryFailure indicates that there was an error when
	// making a PKCS#11 api call.
	ErrQueryFailure = errors.New("unable to execute the requested PKCS#11 api call")
)

var caKeyLabel = "LITE Root CA"
var caKeyUUID = []byte{0xd0, 0x61, 0x9e, 0x62, 0xdd, 0xa2, 0x43, 0xb4, 0xb5, 0x3c, 0x85, 0x0b, 0x07, 0xf0, 0x78, 0x1c}

// An HSM manages a session to a single HSM instance.  Generally,
// there will only be a single of these instances for a given
// execution of the program.
type HSM struct {
	api     *pkcs11.Ctx
	session pkcs11.SessionHandle
	lock    sync.Mutex
}

// NewHSM will attempt to connect to the configured HSM.  It will
// return ErrNoHSM if no HSM has been configured. Other errors
// indicate an actual error connecting to the HSM.  `Close()` should
// be called before exiting, when done with the HSM
func NewHSM() (*HSM, error) {
	// Cleanup function, needed due to lack of errdefer in Go.
	cleanup := func() {}
	module := viper.GetString("hsm.module")
	pin := viper.GetString("hsm.pin")

	if module == "" || pin == "" {
		return nil, ErrNoHSM
	}

	// Make sure the specified PKCS#11 shared module actually
	// exists
	_, err := os.Stat(module)
	if err != nil {
		return nil, ErrInvalidFilename
	}

	// Try to initialize the shared module
	api := pkcs11.New(module)
	if err = api.Initialize(); err != nil {
		return nil, ErrInitFailure
	}
	cleanup = func() {
		api.Finalize()
		api.Destroy()
	}

	// Get the list of slots.
	slots, err := api.GetSlotList(true)
	if err != nil {
		cleanup()
		return nil, ErrMissingSlot
	}

	// For now, just assume that we can use slot 0.  Ideally, we
	// would either let the slot id be passed in, or use a
	// specific name.
	session, err := api.OpenSession(slots[0], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		cleanup()
		return nil, ErrSessionFailure
	}

	// Attempt to log into this slot using the provided PIN.
	if err = api.Login(session, pkcs11.CKU_USER, pin); err != nil {
		cleanup()
		return nil, ErrLoginFailure
	}

	return &HSM{
		api:     api,
		session: session,
	}, nil
}

// Close frees up the resources used by the pkcs 11 session.
func (h *HSM) Close() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	err := h.api.Finalize()

	// Free the resources associated, regardless of whether the
	// finalize was successful or not.
	h.api.Destroy()

	return err
}

// DisplayInfo show information about the current token.
func (h *HSM) DisplayInfo() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	info, err := h.api.GetInfo()

	if err != nil {
		return ErrQueryFailure
	}

	fmt.Printf("HSM:\n")
	fmt.Printf("  Name: %s\n", info.ManufacturerID)
	fmt.Printf("  Cryptoki: %v.%v\n", info.CryptokiVersion.Major, info.CryptokiVersion.Minor)
	fmt.Printf("  Library: %v.%v\n", info.LibraryVersion.Major, info.LibraryVersion.Minor)

	return nil
}

// DisplaySlotInfo shows information about what is in the selected
// slot on the HSM.
func (h *HSM) DisplaySlotInfo() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	sinfo, err := h.api.GetSessionInfo(h.session)
	if err != nil {
		return ErrQueryFailure
	}

	fmt.Printf("Slot 0:\n")
	fmt.Printf("  ID: 0x%x\n", sinfo.SlotID)

	tinfo, err := h.api.GetTokenInfo(sinfo.SlotID)
	if err != nil {
		return ErrQueryFailure
	}

	fmt.Printf("  Token label: %s\n", tinfo.Label)
	fmt.Printf("  Serial: %s\n", tinfo.SerialNumber)

	return nil
}

func (h *HSM) displayObject(o pkcs11.ObjectHandle) error {
	attr, err := h.api.GetAttributeValue(h.session, o, []*pkcs11.Attribute{
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

// FindObjectsByLabel searches the slot on the token for an object
// that matches the specified label, and prints information about it.
func (h *HSM) FindObjectsByLabel(label string, max int) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	// Set the search parameters
	t := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_LABEL, label),
	}

	if err := h.api.FindObjectsInit(h.session, t); err != nil {
		return ErrQueryFailure
	}

	// Get the first object that matches t
	o, _, err := h.api.FindObjects(h.session, max)
	if err != nil {
		return ErrQueryFailure
	}

	// Clean up
	if err = h.api.FindObjectsFinal(h.session); err != nil {
		return ErrQueryFailure
	}

	// Display matching object
	fmt.Printf("Found %d objects with label \"%s\"\n", len(o), label)
	for i, obj := range o {
		fmt.Printf("  Object %d:\n", i)
		if err := h.displayObject(obj); err != nil {
			return err
		}
	}

	return nil
}

// FindObjectsByUUID search the slot on the token for an object with a
// matching uuid to the one specified.
func (h *HSM) FindObjectsByUUID(u uuid.UUID, max int) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	// Set the search parameters
	t := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_ID, []byte(u[0:])),
	}

	if err := h.api.FindObjectsInit(h.session, t); err != nil {
		return ErrQueryFailure
	}

	// Get the first object that matches t
	o, _, err := h.api.FindObjects(h.session, max)
	if err != nil {
		return ErrQueryFailure
	}

	// Clean up
	if err = h.api.FindObjectsFinal(h.session); err != nil {
		return ErrQueryFailure
	}

	// Display matching object
	fmt.Printf("Found %d objects with UUID \"%s\"\n", len(o), u.String())
	for i, obj := range o {
		fmt.Printf("  Object %d:\n", i)
		if err := h.displayObject(obj); err != nil {
			return err
		}
	}

	return nil
}

// Determined with:
// $ openssl ecparam --name prime256v1 -outform DER | hexdump -C
var p256Oid []byte = []byte{0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x03, 0x01, 0x07}

// CreateRootCAKey generates an initial key pair, storing it on the token.
func (h *HSM) CreateRootCAKey() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	// Private key settins
	// t := []*pkcs11.Attribute{
	// 	pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
	// 	pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_EC),
	// 	pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
	// 	pkcs11.NewAttribute(pkcs11.CKA_PRIVATE, true),
	// 	pkcs11.NewAttribute(pkcs11.CKA_SENSITIVE, true),
	// 	pkcs11.NewAttribute(pkcs11.CKA_LABEL, caKeyLabel),
	// 	pkcs11.NewAttribute(pkcs11.CKA_ID, caKeyUUID),
	// 	pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
	// 	pkcs11.NewAttribute(pkcs11.CKA_EXTRACTABLE, false),
	// 	// TODO: Add other required attributes
	// }

	pubAtts := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_VERIFY, true),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_EC_PARAMS, p256Oid),
	}
	privAtts := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
	}

	mech := []*pkcs11.Mechanism{
		pkcs11.NewMechanism(pkcs11.CKM_EC_KEY_PAIR_GEN, nil),
	}

	_, _, err := h.api.GenerateKeyPair(h.session, mech, pubAtts, privAtts)
	if err != nil {
		return ErrQueryFailure
	}

	return nil
}
