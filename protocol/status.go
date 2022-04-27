package protocol // github.com/microbuilder/linaroca/protocol

import "math/big"

type DevStatusResponse struct {
	Status  int       `cbor:"1,keyasint"`
	Serials []big.Int `cbor:"2,keyasint"`
}

type CertStatusResponse struct {
	Status int `cbor:"1,keyasint"`
}
