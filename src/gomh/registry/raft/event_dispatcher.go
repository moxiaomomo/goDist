package raft

import (
	//	"reflect"
	"sync"
)

// 用于管理事件命名及分派事件通知给监听者们
type eventDispatcher struct {
	sync.RWMutex
	source    interface{}
	listeners map[string]eventListeners
}

type EventListener func(Event)

type eventListeners []EventListener

func (d *eventDispatcher) newEventDispatcher(source interface{}) *eventDispatcher {
	//TODO
	return nil
}

func (d *eventDispatcher) AddEventListener(typ string, listener EventListener) {
	//TODO
}

func (d *eventDispatcher) RemoveEventListener(typ string, listener EventListener) {
	//TODO
}

func (d *eventDispatcher) DispatchEvent(e Event) {
	//TODO
}
