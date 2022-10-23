package apmelastic

import (
	"context"
	"fmt"

	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"

	"github.com/skynet2/eventsourcing/publisher"
)

var PublisherElasticApmInterceptor = func() publisher.UnaryPublisherInterceptorFunc {
	return func(next publisher.UnaryPublisherFunc) publisher.UnaryPublisherFunc {
		return func(ctx context.Context, req publisher.AnyEvent) {
			if apmTx := apm.TransactionFromContext(ctx); apmTx != nil {
				req.SetHeader(elasticTraceparentHeader, apmhttp.FormatTraceparentHeader(apmTx.TraceContext()))

				span := apmTx.StartSpan(
					fmt.Sprintf("publish event to %v", req.GetDestination()),
					"eventsourcing",
					apm.SpanFromContext(ctx),
				)

				span.Context.SetLabel("destination", req.GetDestination())
				span.Context.SetDestinationService(apm.DestinationServiceSpanContext{
					Name:     req.GetDestinationType(),
					Resource: req.GetDestination(),
				})

				defer span.End()
			}

			next(ctx, req)
		}
	}
}
