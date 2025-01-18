package publisher

import "github.com/skynet2/eventsourcing/common"

type Serializer[T any] interface {
	Encode(event common.Event[T]) ([]byte, error)
	ContentType() common.ContentType
}
