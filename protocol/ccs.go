package protocol // github.com/Linaro/lite_bootstrap_server/protocol

type CCSResponse struct {
	Hubname string `cbor:"1,keyasint"`
	Port    int    `cbor:"2,keyasint"`
}
