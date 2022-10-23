package consumer

import (
	"context"

	"github.com/skynet2/eventsourcing/pkg/common"
)

type Consumer[T any] interface {
	WithInterceptors(interceptors ...UnaryInterceptorFunc) Consumer[T]
	ConsumeAsync() error
	Close() error
}

type Fn[T any] func(ctx context.Context, event *common.Event[*T]) (ConfirmationType, error)

type MessageRequest interface {
	Header() map[string][]string
	Any() any
}

type ConfirmationType byte

const (
	ConfirmationTypeAck  = ConfirmationType(0)
	ConfirmationTypeNack = ConfirmationType(1)
)

type UnaryInterceptorFunc = func(next UnaryFunc) UnaryFunc
type UnaryFunc = func(ctx context.Context, request MessageRequest) (ConfirmationType, error)
