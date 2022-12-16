package protocol // github.com/Linaro/lite_bootstrap_server/protocol

import "math/big"

// The DevStatusResponse returns information on the device.
type DevStatusResponse struct {
	Status  int       `cbor:"1,keyasint"`
	Serials []big.Int `cbor:"2,keyasint"`
}

// The CertStatusResponse returns status information on the given
// certificate.
type CertStatusResponse struct {
	Status int `cbor:"1,keyasint"`
}
