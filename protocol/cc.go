package protocol // github.com/Linaro/lite_bootstrap_server/protocol

// A CCResponse is the packet response to get a certificate.
type CCResponse struct {
	Status int    `cbor:"1,keyasint"`
	Cert   string `cbor:"2,keyasint"`
}
