package common

const (
	FrameworkVersion = "0.0.1"
)

type Event[T any] struct {
	Record   *T       `json:"r"`
	MetaData MetaData `json:"m"`
}

type MetaData struct {
	CrudOperation       ChangeEvenType `json:"co"`
	CrudOperationReason string         `json:"cor"`
}

type ChangeEvenType int8

const (
	ChangeEventTypeNone    = ChangeEvenType(0)
	ChangeEventTypeCreated = ChangeEvenType(1)
	ChangeEventTypeUpdated = ChangeEvenType(2)
	ChangeEventTypeDeleted = ChangeEvenType(3)
)
