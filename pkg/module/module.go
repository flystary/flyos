// pkg/module/module.go
package module

import "context"

// Module 是所有模块的基础接口
type Module interface {
	Name() string
	Category() string // e.g., "network", "security", "policy"
	Version() string  // e.g., "1.0"
}

// CommandModule: 提供 CLI 命令
type CommandModule interface {
	Module
	RegisterCommands(registry CommandRegistry)
}

// DaemonModule: 后台运行（如 IDS、日志监听）
type DaemonModule interface {
	Module
	Start(ctx context.Context) error
	Stop() error
}

// StatefulModule: 可查询当前状态
type StatefulModule interface {
	Module
	Get(name string) (Spec, error)
	List() ([]Spec, error)
}

// EventHandler: 响应系统事件（用于联动）
type EventHandler interface {
	OnEvent(event Event) Action
}

type CommandRegistry interface {
	Register(cmd string, handler CommandHandler)
}

type CommandHandler func(args []string) error

type Spec map[string]interface{}

type Event struct {
	Type string
	Data map[string]interface{}
}

type Action struct {
	BlockIP bool
	Log     bool
	Alert   bool
	Message string
}

// ModuleObject 统一接口，模块对象实现自己的 Execute
type ModuleObject interface {
	Execute(verb string) error
}
