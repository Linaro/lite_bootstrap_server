package protocol // github.com/microbuilder/linaroca/protocol

type CCResponse struct {
	Status int    `cbor:"1,keyasint"`
	Cert   string `cbor:"2,keyasint"`
}
