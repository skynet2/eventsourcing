package publisher_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"

	"github.com/skynet2/eventsourcing/common"
	"github.com/skynet2/eventsourcing/publisher"
)

func TestNatsPublisher(t *testing.T) {
	con, err := nats.Connect(getNatsUrl())
	assert.NoError(t, err)
	uuid := uuid.NewString()

	type eventStruct struct {
		Text string
	}

	firstInterceptorCalled := false
	secondInterceptorCalled := false

	record := eventStruct{
		Text: uuid,
	}

	meta := common.MetaData{
		CrudOperation:       common.ChangeEventTypeCreated,
		CrudOperationReason: "abcd",
	}

	data, err := json.Marshal(common.Event[eventStruct]{
		Record:   &record,
		MetaData: meta,
	})
	assert.NoError(t, err)

	pub := publisher.NewNatsPublisher[eventStruct](con, uuid,
		func(next publisher.UnaryPublisherFunc) publisher.UnaryPublisherFunc {
			return func(ctx context.Context, event publisher.AnyEvent) {
				assert.True(t, firstInterceptorCalled)
				assert.False(t, secondInterceptorCalled)
				assert.Equal(t, uuid, event.GetDestination())
				assert.Equal(t, getNatsUrl(), event.GetDestinationType())
				assert.Equal(t, data, event.GetBody())

				secondInterceptorCalled = true
				assert.Equal(t, "value1", event.GetHeader("header1")[0])
				assert.Equal(t, "inter1_value", event.GetHeader("inter1_header")[0])
				next(ctx, event)
			}
		}, func(next publisher.UnaryPublisherFunc) publisher.UnaryPublisherFunc {
			return func(ctx context.Context, event publisher.AnyEvent) {
				assert.False(t, firstInterceptorCalled)
				assert.False(t, secondInterceptorCalled)
				firstInterceptorCalled = true
				assert.Equal(t, "value1", event.GetHeader("header1")[0])
				event.SetHeader("inter1_header", "inter1_value")
				assert.Equal(t, data, event.GetBody())
				next(ctx, event)
			}
		})

	assert.NoError(t, pub.Publish(context.TODO(), record, meta, map[string][]string{
		"header1": {"value1"},
	}))
}

func TestPublishWithCancelledContext(t *testing.T) {
	con, err := nats.Connect(getNatsUrl())
	assert.NoError(t, err)
	uuid := uuid.NewString()

	type event struct {
		Text string
	}

	pub := publisher.NewNatsPublisher[event](con, uuid)

	con.Close()
	assert.ErrorContains(t, pub.Publish(context.TODO(), event{
		Text: uuid,
	}, common.MetaData{
		CrudOperation:       common.ChangeEventTypeCreated,
		CrudOperationReason: "abcd",
	}, nil), "nats: connection closed")
}

func TestMarshalInvalid(t *testing.T) {
	pub := publisher.NewNatsPublisher[any](nil, "", nil)
	ch := make(chan int)

	assert.ErrorContains(t, pub.Publish(context.TODO(), ch, common.MetaData{}, nil),
		"json: unsupported type: chan int")
}

func getNatsUrl() string {
	if env := os.Getenv("NATS_HOST"); len(env) > 0 {
		return env
	}

	return nats.DefaultURL
}
