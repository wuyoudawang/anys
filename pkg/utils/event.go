package utils

import (
	"reflect"
)

type Event struct {
	listeners map[string][]reflect.Value
}

func (e *Event) AddEventListener(typ string, i interface{}) {
	if e.listeners == nil {
		e.listeners = make(map[string][]reflect.Value)
	}

	if _, ok := e.listeners[typ]; !ok {
		e.listeners[typ] = make([]reflect.Value, 0)
	}

	if e.IndexOf(typ, i) == -1 {
		e.listeners[typ] = append(e.listeners[typ], reflect.ValueOf(i))
	}
}

func (e *Event) RemoveEventListener(typ string, i interface{}) {
	idx := e.IndexOf(typ, i)
	if idx != -1 {
		for idx < len(e.listeners[typ])-1 {
			e.listeners[typ][idx] = e.listeners[typ][idx+1]
			idx++
		}
		e.listeners[typ] = e.listeners[typ][:len(e.listeners)-1]
	}
}

func (e *Event) IndexOf(typ string, i interface{}) int {
	if e.listeners == nil {
		return -1
	}

	if _, ok := e.listeners[typ]; !ok {
		return -1
	}

	v := reflect.ValueOf(i)
	for idx, cv := range e.listeners[typ] {
		if cv == v {
			return idx
		}
	}
	return -1
}

func (e *Event) DispatchEvent(typ string, args ...interface{}) {
	if e.listeners != nil {
		if set, ok := e.listeners[typ]; ok {
			var in = make([]reflect.Value, len(args))
			for i, arg := range args {
				in[i] = reflect.ValueOf(arg)
			}

			for _, f := range set {
				f.Call(in)
			}
		}
	}
}
