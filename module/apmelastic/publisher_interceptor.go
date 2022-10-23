package apmelastic

import (
	"context"

	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"

	"github.com/skynet2/eventsourcing/publisher"
)

var PublisherElasticApmInterceptor = func() publisher.UnaryPublisherInterceptorFunc {
	return func(next publisher.UnaryPublisherFunc) publisher.UnaryPublisherFunc {
		return func(ctx context.Context, req publisher.AnyEvent) {
			if apmTx := apm.TransactionFromContext(ctx); apmTx != nil {
				req.SetHeader(elasticTraceparentHeader, apmhttp.FormatTraceparentHeader(apmTx.TraceContext()))
			}

			next(ctx, req)
		}
	}
}
