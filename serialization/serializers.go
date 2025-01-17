package serialization

import "github.com/skynet2/eventsourcing/common"

type Encoding interface {
	Encode(record any) ([]byte, error)
	Decode(data []byte, record any) error
	ContentType() common.ContentType
}

func Supported() []Encoding {
	return []Encoding{
		NewJSON(),
		NewProtoJSONEncoder(),
	}
}
