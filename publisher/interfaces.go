package publisher

import "github.com/skynet2/eventsourcing/common"

type Serializer[T any] interface {
	Encode(record any) ([]byte, error)
	ContentType() common.ContentType
}
