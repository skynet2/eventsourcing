package publisher

import (
	"context"

	"github.com/skynet2/eventsourcing/common"
)

type Publisher[T any] interface {
	Publish(
		ctx context.Context,
		record *T,
		meta common.MetaData,
		headers map[string][]string,
	) error
}

type AnyEvent interface {
	SetHeader(header string, value string)
	GetBody() []byte
}

type UnaryPublisherInterceptorFunc = func(next UnaryPublisherFunc) UnaryPublisherFunc
type UnaryPublisherFunc = func(ctx context.Context, event AnyEvent)
