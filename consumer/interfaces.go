package consumer

type Decoder interface {
	Decode(data []byte, result any) error
}
