package protocol // github.com/Linaro/lite_bootstrap_server/protocol

type CSRRequest struct {
	_   struct{} `cbor:",toarray"`
	CSR []byte
}

type CSRResponse struct {
	Status int    `cbor:"1,keyasint"`
	Cert   []byte `cbor:"2,keyasint"`
}
