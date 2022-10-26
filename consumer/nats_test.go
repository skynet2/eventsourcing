package consumer_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"

	"github.com/skynet2/eventsourcing/common"
	"github.com/skynet2/eventsourcing/consumer"
	"github.com/skynet2/eventsourcing/publisher"
)

type eventStruct struct {
	Text string
}

func TestNatsConsumer(t *testing.T) {
	con, err := nats.Connect(getNatsUrl(), nats.Timeout(30*time.Second), nats.ReconnectWait(30*time.Second))
	assert.NoError(t, err)
	js, err := con.JetStream()
	assert.NoError(t, err)
	sub := uuid.NewString()

	_, err = js.AddStream(&nats.StreamConfig{
		Name:        sub,
		Description: "",
		Subjects:    []string{sub},
	}, nats.Context(context.TODO()))
	assert.NoError(t, err)

	_, err = js.AddConsumer(sub, &nats.ConsumerConfig{
		Durable:    sub,
		Name:       sub,
		AckPolicy:  nats.AckExplicitPolicy,
		MaxDeliver: 10,
		AckWait:    100 * time.Second,
	}, nats.Context(context.TODO()))
	assert.NoError(t, err)

	var receivedMessages []common.Event[eventStruct]
	firstInterceptorCalled := false
	secondInterceptorCalled := false

	srv := consumer.NewNatsConsumer[eventStruct](js,
		consumer.NatsConsumerConfiguration{
			Concurrency:  1,
			ConsumerName: sub,
			Stream:       sub,
		},
		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
			receivedMessages = append(receivedMessages, *event)

			return consumer.ConfirmationTypeAck, nil
		}, consumer.WithNatsOptionInterceptors(func(next consumer.UnaryFunc) consumer.UnaryFunc {
			return func(ctx context.Context, request consumer.MessageRequest) (consumer.ConfirmationType, error) {
				if len(receivedMessages) == 0 {
					assert.False(t, firstInterceptorCalled)
					assert.False(t, secondInterceptorCalled)
					firstInterceptorCalled = true
				}

				return next(ctx, request)
			}
		}, func(next consumer.UnaryFunc) consumer.UnaryFunc {
			return func(ctx context.Context, request consumer.MessageRequest) (consumer.ConfirmationType, error) {
				if len(receivedMessages) == 0 {
					assert.True(t, firstInterceptorCalled)
					assert.False(t, secondInterceptorCalled)
					secondInterceptorCalled = true
				}
				return next(ctx, request)
			}
		}))

	assert.NoError(t, srv.ConsumeAsync())

	pub := publisher.NewNatsPublisher[eventStruct](con, sub)

	expected := []common.Event[eventStruct]{
		{
			Record: &eventStruct{Text: "123454321"},
			MetaData: common.MetaData{
				CrudOperation:       common.ChangeEventTypeCreated,
				CrudOperationReason: "created_1234",
			},
		},
		{
			Record: &eventStruct{Text: "321321321"},
			MetaData: common.MetaData{
				CrudOperation:       common.ChangeEventTypeUpdated,
				CrudOperationReason: "updated_12312",
			},
		},
	}

	for _, e := range expected {
		assert.NoError(t, pub.Publish(context.TODO(), *e.Record, e.MetaData, nil))
	}

	time.Sleep(5 * time.Second)

	assert.Len(t, receivedMessages, 2)
	assert.Equal(t, expected, receivedMessages)

	assert.NoError(t, srv.Close())
	assert.NoError(t, srv.Close())
	assert.True(t, firstInterceptorCalled)
	assert.True(t, secondInterceptorCalled)
}
//
//func TestCloseWhileReading(t *testing.T) {
//	con, err := nats.Connect(getNatsUrl(), nats.Timeout(30*time.Second), nats.ReconnectWait(30*time.Second))
//	assert.NoError(t, err)
//	js, err := con.JetStream()
//	assert.NoError(t, err)
//	sub := uuid.NewString()
//
//	_, err = js.AddStream(&nats.StreamConfig{
//		Name:        sub,
//		Description: "",
//		Subjects:    []string{sub},
//	}, nats.Context(context.TODO()))
//	assert.NoError(t, err)
//
//	_, err = js.AddConsumer(sub, &nats.ConsumerConfig{
//		Durable:    sub,
//		Name:       sub,
//		AckPolicy:  nats.AckExplicitPolicy,
//		MaxDeliver: 0,
//		AckWait:    100 * time.Second,
//	}, nats.Context(context.TODO()))
//	assert.NoError(t, err)
//
//	srv := consumer.NewNatsConsumer[eventStruct](js,
//		consumer.NatsConsumerConfiguration{
//			Concurrency:  10,
//			ConsumerName: sub,
//			Stream:       sub,
//		},
//		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
//			return consumer.ConfirmationTypeNack, nil
//		})
//
//	assert.NoError(t, srv.ConsumeAsync())
//
//	pub := publisher.NewNatsPublisher[eventStruct](con, sub)
//
//	for i := 0; i < 100; i++ {
//		assert.NoError(t, pub.Publish(context.TODO(), eventStruct{}, common.MetaData{}, nil))
//	}
//
//	time.Sleep(3 * time.Second)
//
//	assert.NoError(t, srv.Close())
//}
//
//func TestCloseNatsConnection(t *testing.T) {
//	con, err := nats.Connect(getNatsUrl(), nats.Timeout(30*time.Second), nats.ReconnectWait(30*time.Second))
//	assert.NoError(t, err)
//	js, err := con.JetStream()
//	assert.NoError(t, err)
//	sub := uuid.NewString()
//
//	_, err = js.AddStream(&nats.StreamConfig{
//		Name:        sub,
//		Description: "",
//		Subjects:    []string{sub},
//	})
//	assert.NoError(t, err)
//
//	_, err = js.AddConsumer(sub, &nats.ConsumerConfig{
//		Durable:    sub,
//		Name:       sub,
//		AckPolicy:  nats.AckExplicitPolicy,
//		MaxDeliver: 0,
//		AckWait:    100 * time.Second,
//	})
//	assert.NoError(t, err)
//
//	srv := consumer.NewNatsConsumer[eventStruct](js,
//		consumer.NatsConsumerConfiguration{
//			Concurrency:  10,
//			ConsumerName: sub,
//			Stream:       sub,
//		},
//		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
//			return consumer.ConfirmationTypeNack, nil
//		})
//
//	assert.NoError(t, srv.ConsumeAsync())
//
//	pub := publisher.NewNatsPublisher[eventStruct](con, sub)
//
//	for i := 0; i < 100; i++ {
//		assert.NoError(t, pub.Publish(context.TODO(), eventStruct{}, common.MetaData{}, nil))
//	}
//
//	time.Sleep(3 * time.Second)
//	con.Close()
//	time.Sleep(1 * time.Second)
//
//	assert.NoError(t, srv.Close())
//}
//
//func TestOnNonExistingStream(t *testing.T) {
//	con, err := nats.Connect(getNatsUrl(), nats.Timeout(30*time.Second), nats.ReconnectWait(30*time.Second))
//	assert.NoError(t, err)
//	js, err := con.JetStream()
//	assert.NoError(t, err)
//	sub := uuid.NewString()
//
//	srv := consumer.NewNatsConsumer[eventStruct](js,
//		consumer.NatsConsumerConfiguration{
//			Concurrency:  1,
//			ConsumerName: sub,
//			Stream:       sub,
//		},
//		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
//			return consumer.ConfirmationTypeNack, nil
//		})
//
//	assert.NoError(t, srv.ConsumeAsync())
//
//	time.Sleep(17 * time.Second)
//	assert.NoError(t, srv.Close())
//}
//
//func TestCloseNatsDrainConnection(t *testing.T) {
//	con, err := nats.Connect(getNatsUrl(), nats.Timeout(30*time.Second), nats.ReconnectWait(30*time.Second))
//	assert.NoError(t, err)
//	js, err := con.JetStream()
//	assert.NoError(t, err)
//	sub := uuid.NewString()
//
//	_, err = js.AddStream(&nats.StreamConfig{
//		Name:        sub,
//		Description: "",
//		Subjects:    []string{sub},
//	})
//	assert.NoError(t, err)
//
//	_, err = js.AddConsumer(sub, &nats.ConsumerConfig{
//		Durable:    sub,
//		Name:       sub,
//		AckPolicy:  nats.AckExplicitPolicy,
//		MaxDeliver: 0,
//		AckWait:    100 * time.Second,
//	})
//	assert.NoError(t, err)
//
//	srv := consumer.NewNatsConsumer[eventStruct](js,
//		consumer.NatsConsumerConfiguration{
//			Concurrency:  10,
//			ConsumerName: sub,
//			Stream:       sub,
//		},
//		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
//			return consumer.ConfirmationTypeNack, nil
//		})
//
//	assert.NoError(t, srv.ConsumeAsync())
//
//	pub := publisher.NewNatsPublisher[eventStruct](con, sub)
//
//	for i := 0; i < 100; i++ {
//		assert.NoError(t, pub.Publish(context.TODO(), eventStruct{}, common.MetaData{}, nil))
//	}
//
//	time.Sleep(3 * time.Second)
//	assert.NoError(t, con.Drain())
//	time.Sleep(1 * time.Second)
//
//	assert.NoError(t, srv.Close())
//}
//
//func TestNakOnInvalidConfirmationType(t *testing.T) {
//	con, err := nats.Connect(getNatsUrl(), nats.Timeout(30*time.Second), nats.ReconnectWait(30*time.Second))
//	assert.NoError(t, err)
//	js, err := con.JetStream()
//	assert.NoError(t, err)
//	sub := uuid.NewString()
//
//	_, err = js.AddStream(&nats.StreamConfig{
//		Name:        sub,
//		Description: "",
//		Subjects:    []string{sub},
//	})
//	assert.NoError(t, err)
//
//	maxDeliver := 30
//	gotMessages := 0
//	_, err = js.AddConsumer(sub, &nats.ConsumerConfig{
//		Durable:    sub,
//		Name:       sub,
//		AckPolicy:  nats.AckExplicitPolicy,
//		MaxDeliver: maxDeliver,
//		AckWait:    100 * time.Second,
//	})
//	assert.NoError(t, err)
//
//	srv := consumer.NewNatsConsumer[eventStruct](js,
//		consumer.NatsConsumerConfiguration{
//			Concurrency:  1,
//			ConsumerName: sub,
//			Stream:       sub,
//		},
//		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
//			gotMessages += 1
//			return 100, nil
//		})
//
//	assert.NoError(t, srv.ConsumeAsync())
//
//	pub := publisher.NewNatsPublisher[eventStruct](con, sub)
//
//	assert.NoError(t, pub.Publish(context.TODO(), eventStruct{}, common.MetaData{}, nil))
//
//	time.Sleep(5 * time.Second)
//
//	assert.Equal(t, maxDeliver, gotMessages)
//	assert.NoError(t, srv.Close())
//}
//
//func TestNakOnInvalidJsonMessage(t *testing.T) {
//	con, err := nats.Connect(getNatsUrl(), nats.Timeout(30*time.Second), nats.ReconnectWait(30*time.Second))
//	assert.NoError(t, err)
//	js, err := con.JetStream()
//	assert.NoError(t, err)
//	sub := uuid.NewString()
//
//	_, err = js.AddStream(&nats.StreamConfig{
//		Name:        sub,
//		Description: "",
//		Subjects:    []string{sub},
//	})
//	assert.NoError(t, err)
//
//	maxDeliver := 30
//	gotMessages := 0
//	_, err = js.AddConsumer(sub, &nats.ConsumerConfig{
//		Durable:    sub,
//		Name:       sub,
//		AckPolicy:  nats.AckExplicitPolicy,
//		MaxDeliver: maxDeliver,
//		AckWait:    100 * time.Second,
//	})
//	assert.NoError(t, err)
//
//	srv := consumer.NewNatsConsumer[eventStruct](js,
//		consumer.NatsConsumerConfiguration{
//			Concurrency:  1,
//			ConsumerName: sub,
//			Stream:       sub,
//		},
//		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
//			gotMessages += 1
//			return consumer.ConfirmationTypeAck, nil
//		})
//
//	assert.NoError(t, con.Publish(sub, []byte("{")))
//	assert.NoError(t, srv.ConsumeAsync())
//
//	time.Sleep(3 * time.Second)
//
//	assert.Equal(t, 0, gotMessages)
//	assert.NoError(t, srv.Close())
//}
//
//func TestExitOnStreamNotFound(t *testing.T) {
//	sub := uuid.NewString()
//	con, err := nats.Connect(getNatsUrl(), nats.Timeout(30*time.Second), nats.ReconnectWait(30*time.Second))
//	assert.NoError(t, err)
//	js, err := con.JetStream()
//	assert.NoError(t, err)
//
//	srv := consumer.NewNatsConsumer[eventStruct](js,
//		consumer.NatsConsumerConfiguration{
//			Concurrency:  1,
//			ConsumerName: sub,
//			Stream:       sub,
//		},
//		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
//			return consumer.ConfirmationTypeAck, nil
//		}, consumer.WithNatsOptionExitOnStreamNotFound(true))
//
//	assert.ErrorContains(t, srv.ConsumeAsync(), "nats: stream not found")
//	assert.NoError(t, srv.Close())
//}

func getNatsUrl() string {
	if env := os.Getenv("JETSTREAM_HOST"); len(env) > 0 {
		return env
	}

	return nats.DefaultURL
}
