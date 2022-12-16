package protocol // github.com/Linaro/lite_bootstrap_server/protocol

// A CSRRequest is the incoming request for a signature.
type CSRRequest struct {
	_   struct{} `cbor:",toarray"`
	CSR []byte
}

// A CSRResponse is the response to the signature, containing the
// signature itself.
type CSRResponse struct {
	Status int    `cbor:"1,keyasint"`
	Cert   []byte `cbor:"2,keyasint"`
}
