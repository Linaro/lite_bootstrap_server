//go:build never
// +build never

// This program reads the CBOR response to the api/v1/cr request, and
// outputs the retrieved certificate.
//
// Assuming the request was read with
//
//    wget --ca-certificate=SERVER.crt \
//        --post-file USER.cbor \
//        https://localhost:1443/api/v1/cr \
//        -O USER.rsp
//
// This program can be used to extract the certificate:
//
//    go run get_cert_cbor.go -in USER.rsp -out USER.crt

package main

import (
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/microbuilder/linaroca/protocol"
)

var (
	inFile  = flag.String("in", "USER.cbor", "Name of input cbor file")
	outFile = flag.String("out", "USER.crt", "Name of output certificate")
)

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		fmt.Printf("Failure: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	inp, err := os.Open(*inFile)
	if err != nil {
		return err
	}
	defer inp.Close()

	raw, err := ioutil.ReadAll(inp)
	if err != nil {
		return err
	}

	var rsp protocol.CSRResponse
	err = cbor.Unmarshal(raw, &rsp)
	if err != nil {
		return err
	}

	if rsp.Status != 0 {
		return fmt.Errorf("Returned status was not 0")
	}

	var pdata pem.Block
	pdata.Type = "CERTIFICATE"
	pdata.Bytes = rsp.Cert

	outp, err := os.Create(*outFile)
	if err != nil {
		return err
	}
	defer outp.Close()

	err = pem.Encode(outp, &pdata)
	if err != nil {
		return err
	}

	return nil
}
