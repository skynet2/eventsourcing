package consumer

type NatsConsumerConfiguration struct {
	Concurrency  int    `json:"Concurrency"`
	ConsumerName string `json:"ConsumerName"`
	Stream       string `json:"Stream"`
}

type natsMessage struct {
	headers map[string][]string
	request any
	spec    Spec
}

func (n *natsMessage) Header() map[string][]string {
	return n.headers
}

func (n *natsMessage) Any() any {
	return n.request
}

func (n *natsMessage) Spec() Spec {
	return n.spec
}

type natsOptions struct {
	exitOnStreamNotFound bool
}

func WithNatsOptionExitOnStreamNotFound(exitOnStreamNotFound bool) func(opt *natsOptions) {
	return func(opt *natsOptions) {
		opt.exitOnStreamNotFound = exitOnStreamNotFound
	}
}
