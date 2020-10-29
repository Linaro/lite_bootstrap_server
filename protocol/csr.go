package protocol // github.com/microbuilder/linaroca/protocol

type CSRRequest struct {
	CSR []byte
}

type CSRResponse struct {
	Status int
	Cert   []byte
}
