package serialization

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/skynet2/eventsourcing/common"
)

type ProtoJSON[T any] struct{}

func NewProtoJSONEncoder[T any]() *ProtoJSON[T] {
	return &ProtoJSON[T]{}
}

func (j *ProtoJSON[T]) Encode(record any) ([]byte, error) {
	event, ok := record.(protoreflect.ProtoMessage)
	if !ok {
		return nil, errors.Newf("can not cast type %T to protoreflect.ProtoMessage", record)
	}

	return protojson.Marshal(event)
}

func (j *ProtoJSON[T]) Decode(data []byte) (*common.Event[T], error) {
	raw := struct {
		Record   json.RawMessage `json:"r"`
		MetaData common.MetaData `json:"m"`
	}{}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	realEvent := &common.Event[T]{
		Record:   nil,
		MetaData: raw.MetaData,
	}

	var targetData T
	var wrap interface{} = targetData
	cc, ok := wrap.(protoreflect.ProtoMessage)
	if !ok {
		return nil, errors.Newf("can not cast type %T to protoreflect.ProtoMessage", targetData)
	}

	err := protojson.Unmarshal(raw.Record, cc)
	if err != nil {
		return nil, err
	}

	return realEvent, nil
}

func (j *ProtoJSON[T]) ContentType() common.ContentType {
	return common.ContentTypeProtoJSON
}
