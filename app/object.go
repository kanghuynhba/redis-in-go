package main

type ObjectType string

const (
	TypeString ObjectType = "string"
	TypeList   ObjectType = "list"
	TypeStream ObjectType = "stream"
	TypeNone   ObjectType = "none"
)

type Object struct {
	Type ObjectType
	Data any
}

func NewObject(objectType ObjectType, data any) *Object {
	return &Object{
		Type: objectType, Data: data,
	}
}
