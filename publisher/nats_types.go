package publisher

import "github.com/nats-io/nats.go"

type natsEvent struct {
	*nats.Msg
	destinationType string
}

func (n *natsEvent) GetHeader(header string) []string {
	return n.Header.Values(header)
}

func (n *natsEvent) GetDestination() string {
	return n.Subject
}

func (n *natsEvent) GetDestinationType() string {
	return n.destinationType
}

func (n *natsEvent) SetHeader(header string, value string) {
	n.Header.Set(header, value)
}

func (n *natsEvent) GetBody() []byte {
	return n.Data
}
