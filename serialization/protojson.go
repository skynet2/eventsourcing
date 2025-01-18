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

func (j *ProtoJSON[T]) Encode(event common.Event[T]) ([]byte, error) {
	cc, ok := any(event.Record).(protoreflect.ProtoMessage)
	if !ok {
		return nil, errors.Newf("can not cast type %T to protoreflect.ProtoMessage", event.Record)
	}

	recordBytes, err := protojson.Marshal(cc)
	if err != nil {
		return nil, err
	}

	raw := struct {
		Record   json.RawMessage `json:"r"`
		MetaData common.MetaData `json:"m"`
	}{
		Record:   recordBytes,
		MetaData: event.MetaData,
	}

	return json.Marshal(raw)
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
		MetaData: raw.MetaData,
	}

	var targetData T
	cc, ok := any(&targetData).(protoreflect.ProtoMessage)
	if !ok {
		return nil, errors.Newf("can not cast type %T to protoreflect.ProtoMessage", targetData)
	}

	err := protojson.Unmarshal(raw.Record, cc)
	if err != nil {
		return nil, err
	}

	realEvent.Record = &targetData
	return realEvent, nil
}

func (j *ProtoJSON[T]) ContentType() common.ContentType {
	return common.ContentTypeProtoJSON
}
