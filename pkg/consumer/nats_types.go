package consumer

type NatsConsumerConfiguration struct {
	Concurrency  int    `json:"Concurrency"`
	ConsumerName string `json:"ConsumerName"`
	Stream       string `json:"Stream"`
}

type natsMessage struct {
	headers map[string][]string
	request any
}

func (n *natsMessage) Header() map[string][]string {
	return n.headers
}

func (n *natsMessage) Any() any {
	return n.request
}
