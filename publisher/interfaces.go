package publisher

import "github.com/skynet2/eventsourcing/common"

type Serializer interface {
	Encode(record any) ([]byte, error)
	ContentType() common.ContentType
}
