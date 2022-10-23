package consumer

import (
	"context"

	"github.com/skynet2/eventsourcing/common"
)

type Consumer[T any] interface {
	ConsumeAsync() error
	Close() error
}

type Fn[T any] func(ctx context.Context, event *common.Event[*T]) (ConfirmationType, error)

type MessageRequest interface {
	Header() map[string][]string
	Spec() Spec
	Any() any
}

type Spec struct {
	ConsumerName  string
	ConsumerQueue string
	Version       string
}

type ConfirmationType byte

const (
	ConfirmationTypeNack = ConfirmationType(0)
	ConfirmationTypeAck  = ConfirmationType(1)
)

type UnaryInterceptorFunc = func(next UnaryFunc) UnaryFunc
type UnaryFunc = func(ctx context.Context, request MessageRequest) (ConfirmationType, error)
