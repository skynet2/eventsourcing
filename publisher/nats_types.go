package publisher

import "github.com/nats-io/nats.go"

type natsEvent struct {
	*nats.Msg
}

func (n *natsEvent) SetHeader(header string, value string) {
	n.Header.Set(header, value)
}

func (n *natsEvent) GetBody() []byte {
	return n.Data
}
