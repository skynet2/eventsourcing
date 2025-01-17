package serialization

import (
	"encoding/json"

	"github.com/skynet2/eventsourcing/common"
)

type JSON struct {
}

func NewJSON() *JSON {
	return &JSON{}
}

func (j *JSON) Encode(record any) ([]byte, error) {
	return json.Marshal(record)
}

func (j *JSON) Decode(data []byte, record any) error {
	return json.Unmarshal(data, record)
}

func (j *JSON) ContentType() common.ContentType {
	return common.ContentTypeJSON
}
