package apmelastic

import "go.elastic.co/apm/v2"

type serverOptions struct {
	tracer  *apm.Tracer
}