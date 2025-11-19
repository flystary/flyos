package ipc

import (
	"sync"
)

type MessageType string

const (
	MsgRequest  MessageType = "req"
	MsgResponse MessageType = "resp"
	MsgNotify   MessageType = "notify"
)

type Message struct {
	ID      string      `json:"id"`
	Type    MessageType `json:"type"`
	Method  string      `json:"method,omitempty"`
	Payload any         `json:"payload,omitempty"`
	Err     string      `json:"err,omitempty"`
}

type Handler func(payload any) (any, error)

var (
	typeRegistryMu  sync.RWMutex
	typeRegistry    = map[string]any{}
	handlerRegistry = map[string]Handler{}
)

func RegisterType(msgType string, v any) {
	typeRegistryMu.Lock()
	defer typeRegistryMu.Unlock()
	typeRegistry[msgType] = v
}

func GetType(msgType string) any {
	typeRegistryMu.RLock()
	defer typeRegistryMu.RUnlock()
	return typeRegistry[msgType]
}

func RegisterHandler(msgType string, h Handler) {
	typeRegistryMu.Lock()
	defer typeRegistryMu.Unlock()
	handlerRegistry[msgType] = h
}

func GetHandler(msgType string) Handler {
	typeRegistryMu.RLock()
	defer typeRegistryMu.RUnlock()
	return handlerRegistry[msgType]
}

type IPCInterface interface {
	Handlers() map[string]Handler
}

func RegisterModule(m IPCInterface) {
	for msgType, handler := range m.Handlers() {
		RegisterType(msgType, nil)
		RegisterHandler(msgType, handler)
	}
}
