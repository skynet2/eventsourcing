package serialization

import "github.com/skynet2/eventsourcing/common"

type Encoding[T any] interface {
	Encode(event common.Event[T]) ([]byte, error)
	Decode(data []byte) (*common.Event[T], error)
	ContentType() common.ContentType
}

func Supported[T any]() []Encoding[T] {
	return []Encoding[T]{
		NewJSON[T](),
		NewProtoJSONEncoder[T](),
	}
}
