package publisher

import (
	"context"

	"github.com/skynet2/eventsourcing/common"
)

type Publisher[T any] interface {
	Publish(
		ctx context.Context,
		record T,
		meta common.MetaData,
		headers *PublishOptions,
	) error
}

type AnyEvent interface {
	SetHeader(header string, value string)
	GetHeader(header string) []string
	GetBody() []byte
	GetDestination() string
	GetDestinationType() string
}

type event[T any] struct {
	Record   T               `json:"r"`
	MetaData common.MetaData `json:"m"`
}

type PublishOptions struct {
	Headers       map[string][]string
	CustomSubject string
}

type UnaryPublisherInterceptorFunc = func(next UnaryPublisherFunc) UnaryPublisherFunc
type UnaryPublisherFunc = func(ctx context.Context, event AnyEvent)
