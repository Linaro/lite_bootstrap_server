package protocol // github.com/microbuilder/linaroca/protocol

type CSRRequest struct {
	_   struct{} `cbor:",toarray"`
	CSR []byte
}

type CSRResponse struct {
	Status int    `cbor:"1,keyasint"`
	Cert   []byte `cbor:"2,keyasint"`
}

// The ServiceResponse returns information about the intended service
type ServiceResponse struct {
	Status  int    `cbor:"1,keyasint"`
	Hubname string `cbor:"2,keyasint"`
	Port    int    `cbor:"3,keyasint"`
}
