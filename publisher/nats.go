package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"

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
	headers map[string][]string,
) error {
	data, err := json.Marshal(event[T]{
		Record:   record,
		MetaData: meta,
	})

	if err != nil {
		return errors.WithStack(err)
	}

	m := &nats.Msg{
		Subject: n.subject,
		Data:    data,
		Header: map[string][]string{
			"co":  {fmt.Sprint(meta.CrudOperation)},
			"cor": {fmt.Sprint(meta.CrudOperationReason)},
		},
	}

	if len(headers) > 0 {
		for k, v := range headers {
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
