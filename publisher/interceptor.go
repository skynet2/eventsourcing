package publisher

func executeInterceptors(next UnaryClientFunc, mid []UnaryClientInterceptorFunc) UnaryClientFunc {
	for _, interceptor := range mid {
		next = interceptor(next)
	}

	return next
}
