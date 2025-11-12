package routing

import (
	"fmt"
	"sync"
)

var (
	mu       sync.RWMutex
	handlers = make(map[string]func() Route)
)

const (
	defaultEnabled  = true
	defaultDisabled = false
)

func RegisterRoute(proto string, isDefaultEnabled bool, factory func() Route) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := handlers[proto]; exists {
		panic("route type already registered: " + proto)
	}
	handlers[proto] = factory
}

func NewRouteByProto(proto string) (Route, error) {
	mu.RLock()
	factory, ok := handlers[proto]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown route proto: %s", proto)
	}
	return factory(), nil
}

func RouteKey(r Route) string {
	return r.GetProto() + "|" + r.GetPrefix()
}
