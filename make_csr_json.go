// +build never

// This program reads a user CSR key (generated with openssl,
// currently):
//
//    openssl ecparam -name prime256v1 -genkey -out USER.key
//    openssl req -new -key USER.key -out USER.csr \
//       -subj "/O=Orgname/CN=396c7a48-a1a6-4682-ba36-70d13f3b8902"
//
// The CN of the subject of the key should be the uinque identifier
// for the device being simulated.
//
// This program reads that file, and outputs a json csr request
// appropriate to give the server.
//
// This request can be given to the server (assuming port 1443) with
//
//    wget --ca-certificate=SERVER.crt \
//        --post-file USER.json \
//        https://localhost:1443/api/v1/cr

package main

import (
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	inFile  = flag.String("in", "USER.csr", "Name of input csr file")
	outFile = flag.String("out", "USER.json", "Name of output json")
)

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		fmt.Printf("Failure: %v\n", err)
		os.Exit(1)
	}
}

type Request struct {
	CSR []byte
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

	pem, rest := pem.Decode(raw)
	if len(rest) != 0 {
		return fmt.Errorf("Invalid PEM input. Expecting one block")
	}
	if pem.Type != "CERTIFICATE REQUEST" {
		return fmt.Errorf("Expecting BEGIN CERTIFICATE REQUEST")
	}

	req := Request{
		CSR: pem.Bytes,
	}
	encoded, err := json.Marshal(&req)
	if err != nil {
		return nil
	}

	err = ioutil.WriteFile(*outFile, encoded, 0644)
	if err != nil {
		return err
	}

	return nil
}
