package publisher

import (
	"context"

	"github.com/skynet2/eventsourcing/pkg/common"
)

type Publisher[T any] interface {
	Publish(
		ctx context.Context,
		record *T,
		meta common.MetaData,
		headers map[string][]string,
	) error
}
