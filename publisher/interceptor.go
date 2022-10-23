package publisher

func executeInterceptors(next UnaryPublisherFunc, mid []UnaryPublisherInterceptorFunc) UnaryPublisherFunc {
	for _, interceptor := range mid {
		next = interceptor(next)
	}

	return next
}
