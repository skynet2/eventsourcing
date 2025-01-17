package serialization

import (
	"github.com/cockroachdb/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/skynet2/eventsourcing/common"
)

type ProtoJSON struct {
	enc *JSON
}

func NewProtoJSONEncoder() *ProtoJSON {
	return &ProtoJSON{
		enc: NewJSON(),
	}
}

func (j *ProtoJSON) Encode(record any) ([]byte, error) {
	return j.enc.Encode(record)
}

func (j *ProtoJSON) Decode(data []byte, record any) error {
	rec, ok := record.(protoreflect.ProtoMessage)
	if !ok {
		return errors.Newf("record with type %T is not a proto message", record)
	}

	return protojson.Unmarshal(data, rec)
}

func (j *ProtoJSON) ContentType() common.ContentType {
	return common.ContentTypeProtoJSON
}
