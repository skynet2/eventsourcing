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
	con          *nats.Conn
	subject      string
	interceptors []UnaryPublisherInterceptorFunc
	encoders     map[common.ContentType]Serializer[T]
}

func NewNatsPublisher[T any](
	con *nats.Conn,
	subject string,
	interceptors ...UnaryPublisherInterceptorFunc,
) Publisher[T] {
	publisher := &NatsPublisher[T]{
		con:          con,
		subject:      subject,
		interceptors: interceptors,
		encoders:     make(map[common.ContentType]Serializer[T]),
	}

	for _, enc := range serialization.Supported[T]() {
		publisher.encoders[enc.ContentType()] = enc
	}

	return publisher
}

func (n *NatsPublisher[T]) Publish(
	ctx context.Context,
	record T,
	meta common.MetaData,
	publishOptions *PublishOptions,
) error {
	encoder := n.encoders[common.ContentTypeJSON]

	if publishOptions != nil && len(publishOptions.Headers[common.ContentTypeHeader]) > 0 {
		encoder = n.encoders[common.ContentType(publishOptions.Headers[common.ContentTypeHeader][0])]
	}

	data, err := encoder.Encode(common.Event[T]{
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
			"co":  {fmt.Sprint(meta.CrudOperation)},
			"cor": {fmt.Sprint(meta.CrudOperationReason)},
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
	}, n.interceptors)(ctx, &natsEvent{
		Msg:             m,
		destinationType: n.con.ConnectedUrl(),
	})

	return err
}
