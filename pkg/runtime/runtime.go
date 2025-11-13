// pkg/runtime/runtime.go
package runtime

import (
	"context"
	"fmt"
	"sync"

	"flyos/internal/converter"
	"flyos/pkg/module"
)

// Runtime 管理模块、命令、事件、转换器
type Runtime struct {
	ctx        context.Context
	cancel     context.CancelFunc
	modules    map[string]module.Module         // 已注册模块
	commands   map[string]module.CommandHandler // 命令注册表
	converters map[string]converter.Converter   // kind -> converter
	eventBus   *EventBus
	mu         sync.RWMutex
}

// New 创建 Runtime
func New() *Runtime {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runtime{
		ctx:        ctx,
		cancel:     cancel,
		modules:    make(map[string]module.Module),
		commands:   make(map[string]module.CommandHandler),
		converters: make(map[string]converter.Converter),
		eventBus:   NewEventBus(),
	}
}

// 注册模块
func (rt *Runtime) RegisterModule(m module.Module) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.modules[m.Name()] = m

	// 注册命令
	if cmdMod, ok := m.(module.CommandModule); ok {
		cmdMod.RegisterCommands(rt)
	}

	// 启动守护进程
	if daemon, ok := m.(module.DaemonModule); ok {
		go func() {
			if err := daemon.Start(rt.ctx); err != nil {
				fmt.Printf("[ERROR] Module %s daemon failed: %v\n", m.Name(), err)
			}
		}()
	}

	// 订阅事件
	if handler, ok := m.(module.EventHandler); ok {
		rt.eventBus.Subscribe(m.Name(), func(e module.Event) {
			handler.OnEvent(e)
		})
	}
}

// 注册命令
func (rt *Runtime) Register(cmd string, handler module.CommandHandler) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.commands[cmd] = handler
}

// 注册 Converter
func (rt *Runtime) RegisterConverter(kind string, c converter.Converter) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.converters[kind] = c
}

// 执行通用命令（REST/MCP）
func (rt *Runtime) ExecuteCommand(cmd string, args []string) error {
	rt.mu.RLock()
	handler, exists := rt.commands[cmd]
	rt.mu.RUnlock()

	if !exists {
		return fmt.Errorf("unknown command: %s", cmd)
	}

	return handler(args)
}

// 执行 DSL Command
func (rt *Runtime) ExecuteDSLCommand(c interface{}) error {
	// 这里 c 可以是 flyos/pkg/dsl.Command
	if dslCmd, ok := c.(interface {
		Kind() string
		Verb() string
		Args() []string
	}); ok {
		return rt.ExecuteCommand(dslCmd.Kind(), dslCmd.Args())
	}
	return fmt.Errorf("unsupported DSL command type")
}

// 执行 REST/MCP JSON Command
func (rt *Runtime) ExecuteFromJSON(kind string, data map[string]interface{}) error {
	rt.mu.RLock()
	converter, ok := rt.converters[kind]
	rt.mu.RUnlock()
	if !ok {
		return fmt.Errorf("no converter registered for kind '%s'", kind)
	}

	// Converter 将 JSON → ModuleObject
	moduleObj, err := converter.ConvertFromJSON(data)
	if err != nil {
		return fmt.Errorf("convert failed: %w", err)
	}

	// ModuleObject 内部实现 Execute(verb)
	execObj, ok := moduleObj.(module.ModuleObject)
	if !ok {
		return fmt.Errorf("unsupported module type: %T", moduleObj)
	}

	verb := ""
	if v, ok := data["verb"].(string); ok {
		verb = v
	}
	return execObj.Execute(verb)
}

// 发布事件
func (rt *Runtime) PublishEvent(typ string, data map[string]interface{}) {
	rt.eventBus.Publish(module.Event{Type: typ, Data: data})
}

// 启动 Runtime（简化）
func (rt *Runtime) Start() {
	fmt.Println("FlyOS Runtime started. Modules & commands:")
	for name := range rt.modules {
		fmt.Printf("  - Module: %s\n", name)
	}
	for cmd := range rt.commands {
		fmt.Printf("  - Command: %s\n", cmd)
	}
	fmt.Println("Ready to accept commands (DSL / REST / MCP)")
}

// 停止 Runtime
func (rt *Runtime) Stop() {
	rt.cancel()
	// TODO: 停止所有 DaemonModule
}
