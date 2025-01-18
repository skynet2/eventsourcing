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

func (j *JSON[T]) Encode(record any) ([]byte, error) {
	return json.Marshal(record)
}

func (j *JSON[T]) Decode(data []byte) (*common.Event[T], error) {
	var targetStruct common.Event[T]
	if err := json.Unmarshal(data, &targetStruct); err != nil {
		return nil, err
	}

	return &targetStruct, nil
}

func (j *JSON[T]) ContentType() common.ContentType {
	return common.ContentTypeJSON
}
