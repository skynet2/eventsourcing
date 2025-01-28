package publisher

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nats-io/nats.go"

	"github.com/skynet2/eventsourcing/common"
	"github.com/skynet2/eventsourcing/serialization"
)

type NatsPublisher[T any] struct {
	con     *nats.Conn
	subject string
	opts    *publishOptions[T]
}

type publishOptions[T any] struct {
	interceptors []UnaryPublisherInterceptorFunc
	serializer   Serializer[T]
}

type OptionFn[T any] func(*publishOptions[T])

func WithInterceptors[T any](interceptors ...UnaryPublisherInterceptorFunc) OptionFn[T] {
	return func(o *publishOptions[T]) {
		o.interceptors = interceptors
	}
}

func WithSerializer[T any](serializer Serializer[T]) OptionFn[T] {
	return func(o *publishOptions[T]) {
		o.serializer = serializer
	}
}

func NewNatsPublisher[T any](
	con *nats.Conn,
	subject string,
	options ...OptionFn[T],
) Publisher[T] {
	publisher := &NatsPublisher[T]{
		con:     con,
		subject: subject,
	}

	defaultOpt := &publishOptions[T]{
		interceptors: []UnaryPublisherInterceptorFunc{},
		serializer:   serialization.NewJSON[T](),
	}
	for _, fn := range options {
		fn(defaultOpt)
	}

	publisher.opts = defaultOpt

	return publisher
}

func (n *NatsPublisher[T]) Publish(
	ctx context.Context,
	record T,
	meta common.MetaData,
	publishOptions *PublishOptions,
) error {
	data, err := n.opts.serializer.Marshal(common.Event[T]{
		Record:   &record,
		MetaData: meta,
	})

	if err != nil {
		return errors.WithStack(err)
	}

	subject := n.subject
	if publishOptions != nil && publishOptions.CustomSubject != "" {
		subject = publishOptions.CustomSubject
	}

	m := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header: map[string][]string{
			"co":                     {fmt.Sprint(meta.CrudOperation)},
			"cor":                    {fmt.Sprint(meta.CrudOperationReason)},
			common.ContentTypeHeader: {string(n.opts.serializer.ContentType())},
		},
	}

	if publishOptions != nil && len(publishOptions.Headers) > 0 {
		for k, v := range publishOptions.Headers {
			m.Header[k] = v
		}
	}

	executeInterceptors(func(ctx context.Context, request AnyEvent) {
		err = n.con.PublishMsg(m)

		if err != nil {
			err = errors.WithStack(err)
		}
	}, n.opts.interceptors)(ctx, &natsEvent{
		Msg:             m,
		destinationType: n.con.ConnectedUrl(),
	})

	return err
}
