package protocol // github.com/microbuilder/linaroca/protocol

type CCSResponse struct {
	Hubname string `cbor:"1,keyasint"`
	Port    int    `cbor:"2,keyasint"`
}
