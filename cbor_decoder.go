//go:build never
// +build never

// Utility program helps to decode the cbor encoded binary output, focused
// mainly decode curl command with Rest API request with CBOR format response,
// refer to the readme for more information.
//
// This program expects two command-line argument
// 	-`-i filename`: CBOR encoded binary input file.
// 	-`-r resFmt`: Supported formats "cc", "cs", "ds", "csr" ,"ccs.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fxamacker/cbor/v2"
	"github.com/Linaro/lite_bootstrap_server/protocol"
	"io/ioutil"
	"os"
)

var (
	inFile = flag.String("i", "cbor_enc.raw", "Input CBOR encoded file in binary format")
	resFmt = flag.String("r", "", "Expected protocol response format")
)

func main() {
	flag.Parse()

	// Use switch on the resFmt variable.
	switch *resFmt {
	case "cc":
		var ccResponse protocol.CCResponse
		err := decode(&ccResponse)
		if err != nil {
			fmt.Printf("Failure: %v\n", err)
			goto ERROR
		}
	case "cs":
		var csResponse protocol.CertStatusResponse
		err := decode(&csResponse)
		if err != nil {
			fmt.Printf("Failure: %v\n", err)
			goto ERROR
		}
	case "ds":
		var dsResponse protocol.DevStatusResponse
		err := decode(&dsResponse)
		if err != nil {
			fmt.Printf("Failure: %v\n", err)
			goto ERROR
		}
	case "csr":
		var csrResponse protocol.CSRResponse
		err := decode(&csrResponse)
		if err != nil {
			fmt.Printf("Failure: %v\n", err)
			goto ERROR
		}
	case "ccs":
		var ccsResponse protocol.CCSResponse
		err := decode(&ccsResponse)
		if err != nil {
			fmt.Printf("Failure: %v\n", err)
			goto ERROR
		}
	case "":
		fmt.Println("Missing protocol response format argument")
		goto ERROR
	}

	return
ERROR:
	os.Exit(1)

}

func decode(v interface{}) error {
	inp, err := os.Open(*inFile)
	if err != nil {
		return err
	}
	defer inp.Close()

	raw, err := ioutil.ReadAll(inp)
	if err != nil {
		return err
	}

	// Decode to get response struct
	dec := cbor.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(v); err != nil {
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println(err)
	}

	// Convert struct to json format
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Printf("%s\n", data)

	return nil
}
