package protocol // github.com/Linaro/lite_bootstrap_server/protocol

import "math/big"

type DevStatusResponse struct {
	Status  int       `cbor:"1,keyasint"`
	Serials []big.Int `cbor:"2,keyasint"`
}

type CertStatusResponse struct {
	Status int `cbor:"1,keyasint"`
}
