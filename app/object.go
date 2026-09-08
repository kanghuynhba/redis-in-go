package main

import (
	"time"
)

type ObjectType string

const (
	TypeString ObjectType = "string"
	TypeList   ObjectType = "list"
	TypeStream ObjectType = "stream"
	TypeNone   ObjectType = "none"
)

type Object struct {
	Type      ObjectType
	Data      any
	ExpiredAt *time.Time
}
