package middleware

import (
	"sync/atomic"
)

type MwId int32 // 中间件类型ID
type QId int32  // 队列ID
type AId int32  // 身份ID
type Action uint32

type Mwer interface {
	Bind(q string, a Action, c interface{}) QId
	Produce(id QId, message interface{}) interface{}
	GetQId(q string) QId
	Release(q string, c interface{})
}

type Consumer interface {
	Consume(m interface{}) interface{}
}

// 中间件类型

// 行为
const (
	A_PRODUCE Action = 1 + iota
	A_CONSUME
)

type newCp func() Mwer

var components map[MwId]newCp
var mwId MwId = 1000

func init() {
	components = make(map[MwId]newCp, 64)
}

func NewMwId() MwId {
	return MwId(atomic.AddInt32((*int32)(&mwId), 8))
}

func Register(id MwId, New newCp) {
	components[id] = New
}
