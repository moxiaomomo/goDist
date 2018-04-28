package raft

// 监听的事件类型
type Event interface {
	Type() string
	Source() interface{}
	Value() interface{}
	PreValue() interface{}
}

// 接口Event类型的具体实现
type event struct {
	typ      string
	source   interface{}
	value    interface{}
	prevalue interface{}
}

func (e *event) Type() string {
	return e.typ
}

func (e *event) Source() interface{} {
	return e.source
}

func (e *event) Value() interface{} {
	return e.value
}

func (e *event) PreValue() interface{} {
	return e.prevalue
}
