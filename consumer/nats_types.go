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
	interceptors         []UnaryInterceptorFunc
}

func WithNatsOptionExitOnStreamNotFound(exitOnStreamNotFound bool) func(opt *natsOptions) {
	return func(opt *natsOptions) {
		opt.exitOnStreamNotFound = exitOnStreamNotFound
	}
}

func WithNatsOptionInterceptors(interceptors ...UnaryInterceptorFunc) func(opt *natsOptions) {
	return func(opt *natsOptions) {
		opt.interceptors = nil

		for i := len(interceptors) - 1; i >= 0; i-- {
			if interceptor := interceptors[i]; interceptor != nil {
				opt.interceptors = append(opt.interceptors, interceptor)
			}
		}
	}
}
