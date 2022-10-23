package consumer

func executeInterceptors(next UnaryFunc, mid []UnaryInterceptorFunc) UnaryFunc {
	for _, interceptor := range mid {
		next = interceptor(next)
	}

	return next
}
