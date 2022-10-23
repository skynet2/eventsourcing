package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gammazero/workerpool"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/skynet2/eventsourcing/pkg/common"
)

type NatsConsumer[T any] struct {
	wPool         *workerpool.WorkerPool
	jetStream     nats.JetStreamContext
	subscriptions []*nats.Subscription
	cfg           NatsConsumerConfiguration
	isClosing     bool
	fn            Fn[T]
	interceptors  []UnaryInterceptorFunc
	logger        zerolog.Logger
	closeMut      sync.Mutex
}

func NewNatsConsumer[T any](
	natsJetStream nats.JetStreamContext,
	cfg NatsConsumerConfiguration,
	fn Fn[T],
) Consumer[T] {
	return &NatsConsumer[T]{
		jetStream: natsJetStream,
		wPool:     workerpool.New(cfg.Concurrency),
		cfg:       cfg,
		fn:        fn,
		logger:    log.Logger,
	}
}

func (n *NatsConsumer[T]) ConsumeAsync() error {
	for i := 0; i < n.cfg.Concurrency; i++ {
		subscription, err := n.jetStream.PullSubscribe(
			n.cfg.Stream,
			"",
			nats.Bind(n.cfg.Stream, n.cfg.ConsumerName),
		)

		if err != nil {
			return errors.WithStack(err)
		}

		n.subscriptions = append(n.subscriptions, subscription)

		n.wPool.Submit(func() {
			for {
				msg, err := subscription.Fetch(1)

				if n.isClosing {
					if len(msg) > 0 {
						_ = msg[0].Nak()
					}

					return // we can ignore everything, as its no longer important here
				}

				if err != nil {
					if errors.Is(err, nats.ErrConnectionClosed) {
						go func() {
							_ = n.Close() // avoid deadlock
						}()

						return
					}

					if errors.Is(err, nats.ErrTimeout) {
						continue
					}

					n.logger.Err(errors.Wrap(err, fmt.Sprintf("unhendled error from nats: %+v", err))).Send()

					continue
				}

				if len(msg) == 0 { // should not happen
					continue
				}

				targetMsg := msg[0]

				ctx, cancel := context.WithCancel(context.Background())

				confirmationType, err := executeInterceptors(func(ctx context.Context, request MessageRequest) (ConfirmationType, error) { //nolint
					var targetStruct common.Event[*T]

					if err2 := json.Unmarshal(targetMsg.Data, &targetStruct); err2 != nil {
						return ConfirmationTypeNack, err2
					}

					return n.fn(ctx, &targetStruct)
				}, n.interceptors)(ctx, &natsMessage{
					headers: targetMsg.Header,
					request: targetMsg.Data,
				})

				switch confirmationType {
				case ConfirmationTypeAck:
					if respErr := targetMsg.Ack(); respErr != nil {
						n.logger.Err(errors.Wrap(respErr, "can not ack message for default")).Send()
					}
				case ConfirmationTypeNack:
					if respErr := targetMsg.Nak(); respErr != nil {
						n.logger.Err(errors.Wrap(respErr, "can not nack message")).Send()
					}
				default:
					if respErr := targetMsg.Nak(); respErr != nil {
						n.logger.Err(errors.Wrap(respErr, "can not nack message for default")).Send()
					}

					n.logger.Err(errors.New(fmt.Sprintf("unsupported confirmation type %v", confirmationType))).
						Send()
				}

				cancel()
			}
		})
	}

	return nil
}

func (n *NatsConsumer[T]) WithInterceptors(interceptors ...UnaryInterceptorFunc) Consumer[T] {
	n.interceptors = nil

	for i := len(interceptors) - 1; i >= 0; i-- {
		if interceptor := interceptors[i]; interceptor != nil {
			n.interceptors = append(n.interceptors, interceptor)
		}
	}

	return n
}

func (n *NatsConsumer[T]) Close() error {
	n.closeMut.Lock()
	defer n.closeMut.Unlock()

	if n.isClosing {
		return nil
	}

	var finalErr error

	if n.wPool != nil {
		n.wPool.StopWait()
	}

	return finalErr
}
