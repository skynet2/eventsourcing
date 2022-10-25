package consumer_test

//import (
//	"context"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/nats-io/nats.go"
//	"github.com/stretchr/testify/assert"
//
//	"github.com/skynet2/eventsourcing/common"
//	"github.com/skynet2/eventsourcing/consumer"
//)
//
//func TestNatsConsumer(t *testing.T) {
//	con, err := nats.Connect(getNatsUrl())
//	assert.NoError(t, err)
//	js, err := con.JetStream()
//	assert.NoError(t, err)
//
//	type eventStruct struct {
//		Text string
//	}
//
//	srv := consumer.NewNatsConsumer[eventStruct](js,
//		consumer.NatsConsumerConfiguration{
//			Concurrency: 1,
//		},
//		func(ctx context.Context, event *common.Event[eventStruct]) (consumer.ConfirmationType, error) {
//			return consumer.ConfirmationTypeAck, nil
//		})
//
//	assert.NoError(t, srv.ConsumeAsync())
//	time.Sleep(2 * time.Second)
//}
//
//func getNatsUrl() string {
//	if env := os.Getenv("NATS_HOST"); len(env) > 0 {
//		return env
//	}
//
//	return nats.DefaultURL
//}
