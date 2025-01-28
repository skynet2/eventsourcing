package serialization

import (
	"encoding/json"

	"github.com/skynet2/eventsourcing/common"
)

type JSON[T any] struct {
}

func NewJSON[T any]() *JSON[T] {
	return &JSON[T]{}
}

func (j *JSON[T]) Marshal(event common.Event[T]) ([]byte, error) {
	return json.Marshal(event)
}

func (j *JSON[T]) Unmarshal(data []byte) (*common.Event[T], error) {
	var targetStruct common.Event[T]
	if err := json.Unmarshal(data, &targetStruct); err != nil {
		return nil, err
	}

	return &targetStruct, nil
}

func (j *JSON[T]) ContentType() common.ContentType {
	return common.ContentTypeJSON
}
