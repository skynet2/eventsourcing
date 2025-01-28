package consumer

import "github.com/skynet2/eventsourcing/common"

type Decoder[T any] interface {
	Unmarshal(data []byte) (*common.Event[T], error)
}
