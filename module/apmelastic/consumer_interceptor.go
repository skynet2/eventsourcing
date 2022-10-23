package apmelastic

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"go.elastic.co/apm/module/apmhttp/v2"

	"go.elastic.co/apm/v2"

	"github.com/skynet2/eventsourcing/consumer"
)

const (
	frameworkName = "skynet2/eventsourcing"
)

var ConsumerElasticApmInterceptor = func(captureErrors bool) consumer.UnaryInterceptorFunc {
	opts := serverOptions{
		tracer: apm.DefaultTracer(),
	}

	return func(next consumer.UnaryFunc) consumer.UnaryFunc {
		return func(ctx context.Context, req consumer.MessageRequest) (resp consumer.ConfirmationType, err error) {
			if !opts.tracer.Recording() {
				return next(ctx, req)
			}

			spec := req.Spec()
			requestName := fmt.Sprintf("%v with stream %v", spec.ConsumerName, spec.ConsumerQueue)

			tx, ctx := startTransaction(ctx, opts.tracer, requestName, req.Header(), spec.Version)
			defer tx.End()

			defer func() {
				r := recover()
				if r != nil {
					e := opts.tracer.Recovered(r)
					e.SetTransaction(tx)
					tx.Context.SetFramework(frameworkName, spec.Version)
					e.Handled = false
					e.Send()

					panic(r)
				}

				setTransactionResult(tx, resp)

				if err != nil && captureErrors {
					logError(err, ctx)
				}
			}()

			resp, err = next(ctx, req)

			return resp, err
		}
	}
}

func logError(err error, ctx context.Context) {
	if err == nil {
		return
	}

	defer func() {
		_ = recover()
	}()

	log.Ctx(ctx).Err(err).Send()

	apmError := apm.CaptureError(ctx, err)

	if err != nil {
		apmError.Send()
	}
}

func getIncomingMetadataTraceContext(md http.Header, header string) (apm.TraceContext, bool) {
	if values := md.Get(header); len(values) > 0 {
		traceContext, err := apmhttp.ParseTraceparentHeader(values)
		if err == nil {
			traceContext.State, _ = apmhttp.ParseTracestateHeader([]string{md.Get(header)}...)
			return traceContext, true
		}
	}

	return apm.TraceContext{}, false
}

var (
	elasticTraceparentHeader = strings.ToLower(apmhttp.ElasticTraceparentHeader)
	w3cTraceparentHeader     = strings.ToLower(apmhttp.W3CTraceparentHeader)
	tracestateHeader         = strings.ToLower(apmhttp.TracestateHeader)
)

func startTransaction(
	ctx context.Context,
	tracer *apm.Tracer,
	name string,
	header map[string][]string,
	version string,
) (*apm.Transaction, context.Context) {
	var opts apm.TransactionOptions

	traceContext, ok := getIncomingMetadataTraceContext(header, w3cTraceparentHeader)
	if !ok {
		traceContext, ok = getIncomingMetadataTraceContext(header, elasticTraceparentHeader)
	}
	if !ok {
		traceContext, _ = getIncomingMetadataTraceContext(header, tracestateHeader)
	}

	opts.TraceContext = traceContext

	tx := tracer.StartTransactionOptions(name, "request", opts)
	tx.Context.SetFramework(frameworkName, version)

	return tx, apm.ContextWithTransaction(ctx, tx)
}

func setTransactionResult(tx *apm.Transaction, resp consumer.ConfirmationType) {
	tx.Outcome = tx.Result

	if resp == consumer.ConfirmationTypeAck {
		tx.Result = "success"
	} else {
		tx.Result = "failure"
	}
}
