// pkg/runtime/eventbus.go
package runtime

import (
	"flyos/pkg/module"
	"sync"
)

type EventBus struct {
	listeners map[string][]func(module.Event)
	mu        sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]func(module.Event)),
	}
}

func (eb *EventBus) Subscribe(topic string, fn func(module.Event)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.listeners[topic] = append(eb.listeners[topic], fn)
}

func (eb *EventBus) Publish(event module.Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// 广播给所有订阅者（实际可按 event.Type 路由）
	for topic, fns := range eb.listeners {
		for _, fn := range fns {
			go fn(event) // 异步处理
		}
	}
}
