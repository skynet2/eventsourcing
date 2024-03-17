package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nats-io/nats.go"

	"github.com/skynet2/eventsourcing/common"
)

type NatsPublisher[T any] struct {
	con          *nats.Conn
	subject      string
	interceptors []UnaryPublisherInterceptorFunc
}

func NewNatsPublisher[T any](
	con *nats.Conn,
	subject string,
	interceptors ...UnaryPublisherInterceptorFunc,
) Publisher[T] {
	return &NatsPublisher[T]{
		con:          con,
		subject:      subject,
		interceptors: interceptors,
	}
}

func (n *NatsPublisher[T]) Publish(
	ctx context.Context,
	record T,
	meta common.MetaData,
	publishOptions *PublishOptions,
) error {
	data, err := json.Marshal(event[T]{
		Record:   record,
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
