package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

// Command 接口
type Command interface {
	Name() string
	Path() string // 外部命令路径，内置命令返回 ""
	Execute(args []string, env map[string]string) error
	IsBuiltin() bool
	Desc() string
	Usage() string
	Args() []string
	Returns() []string
	Flags() []string
	Subcommands() []string
	Category() string
}

// External File
type FileCommand struct {
	name     string
	path     string
	category string
}

func (f *FileCommand) Name() string          { return f.name }
func (f *FileCommand) Category() string      { return f.category }
func (f *FileCommand) Path() string          { return f.path }
func (f *FileCommand) IsBuiltin() bool       { return false }
func (f *FileCommand) Desc() string          { return "" }
func (f *FileCommand) Usage() string         { return "" }
func (f *FileCommand) Args() []string        { return nil }
func (f *FileCommand) Returns() []string     { return nil }
func (f *FileCommand) Flags() []string       { return nil }
func (f *FileCommand) Subcommands() []string { return nil }
func (f *FileCommand) Execute(args []string, env map[string]string) error {
	cmd := exec.Command(f.path, args[1:]...)
	cmd.Env = mergeEnv(env)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Builtin List
type ListCommand struct{}

func (l *ListCommand) Name() string          { return "list" }
func (l *ListCommand) Category() string      { return "sys" }
func (l *ListCommand) Path() string          { return "" }
func (l *ListCommand) IsBuiltin() bool       { return true }
func (l *ListCommand) Desc() string          { return "打印全部命令" }
func (l *ListCommand) Usage() string         { return "exit" }
func (l *ListCommand) Args() []string        { return []string{""} }
func (l *ListCommand) Returns() []string     { return []string{"展示所有命令！"} }
func (l *ListCommand) Flags() []string       { return nil }
func (l *ListCommand) Subcommands() []string { return nil }
func (l *ListCommand) Execute(args []string, env map[string]string) error {

	return nil
}

// Builtin Exit
type ExitCommand struct{}

func (e *ExitCommand) Name() string          { return "exit" }
func (e *ExitCommand) Category() string      { return "sys" }
func (e *ExitCommand) Path() string          { return "" }
func (e *ExitCommand) IsBuiltin() bool       { return true }
func (e *ExitCommand) Desc() string          { return "退出flyos环境" }
func (e *ExitCommand) Usage() string         { return "exit" }
func (e *ExitCommand) Args() []string        { return []string{"[]"} }
func (e *ExitCommand) Returns() []string     { return []string{"退出环境！"} }
func (e *ExitCommand) Flags() []string       { return nil }
func (e *ExitCommand) Subcommands() []string { return nil }
func (e *ExitCommand) Execute(args []string, env map[string]string) error {
	fmt.Println("👋 Bye!")
	return nil
}

// Builtin Env
type EnvCommand struct{}

func (e *EnvCommand) Name() string          { return "env" }
func (e *EnvCommand) Category() string      { return "sys" }
func (e *EnvCommand) Path() string          { return "" }
func (e *EnvCommand) IsBuiltin() bool       { return true }
func (e *EnvCommand) Desc() string          { return "打印环境变量" }
func (e *EnvCommand) Usage() string         { return "env [VAR...]" }
func (e *EnvCommand) Args() []string        { return []string{"VAR 可选，需要打印的环境变量"} }
func (e *EnvCommand) Returns() []string     { return []string{"打印环境变量内容"} }
func (e *EnvCommand) Flags() []string       { return nil }
func (e *EnvCommand) Subcommands() []string { return nil }
func (e *EnvCommand) Execute(args []string, env map[string]string) error {
	allEnv := mergeEnv(env)
	if len(args) <= 1 {
		for _, v := range allEnv {
			fmt.Println(v)
		}
		return nil
	}
	filter := make(map[string]bool)
	for _, key := range args[1:] {
		filter[key] = true
	}
	for _, v := range allEnv {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if filter[parts[0]] {
			fmt.Println(v)
		}
	}
	return nil
}

// Builtin Help
type HelpCommand struct {
	descMgr *DescManager
	shell   *Shell
}

func NewHelpCommand(d *DescManager, s *Shell) *HelpCommand {
	return &HelpCommand{descMgr: d, shell: s}
}

func (h *HelpCommand) Name() string     { return "help" }
func (h *HelpCommand) Category() string { return "sys" }
func (h *HelpCommand) Path() string     { return "" }
func (h *HelpCommand) IsBuiltin() bool  { return true }
func (h *HelpCommand) Desc() string {
	return "显示命令或分类的帮助信息（支持模糊搜索）"
}
func (h *HelpCommand) Usage() string { return "help [COMMAND|CATEGORY|KEYWORD]" }
func (h *HelpCommand) Args() []string {
	return []string{"命令名、分类名或关键字（可选）"}
}
func (h *HelpCommand) Returns() []string     { return []string{"打印帮助信息"} }
func (h *HelpCommand) Flags() []string       { return nil }
func (h *HelpCommand) Subcommands() []string { return nil }

func (h *HelpCommand) Execute(args []string, env map[string]string) error {
	// 只输入 help 时
	builtins := []Command{}
	if len(args) == 1 {
		// 1️⃣ 打印所有内置命令
		fmt.Println("📦 内置命令:")
		for _, cmd := range h.shell.commands {
			if cmd.IsBuiltin() {
				builtins = append(builtins, cmd)
			}
		}
		sort.Slice(builtins, func(i, j int) bool { return builtins[i].Name() < builtins[j].Name() })
		for _, c := range builtins {
			fmt.Printf("  %-10s - %s\n", c.Name(), c.Desc())
		}
		fmt.Println()

		// 2️⃣ 打印外部命令分类
		cats := h.descMgr.getAllCategories()
		if len(cats) > 0 {
			fmt.Println("📂 外部命令分类:")
			sort.Strings(cats)
			for _, c := range cats {
				fmt.Println("  " + c)
			}
			fmt.Println("\n使用 `help [分类名]` 查看分类内命令")
		} else {
			fmt.Println("⚠️ 暂无外部命令")
		}

		return nil
	}

	// 以下为带参数时的原有逻辑
	target := args[1]

	// 精确匹配内置命令
	if cmd, ok := h.shell.commands[target]; ok && cmd.IsBuiltin() {
		h.descMgr.PrintHelp(target, h.shell)
		return nil
	}

	// 精确匹配外部命令
	if _, ok := h.descMgr.Get(target); ok {
		h.descMgr.PrintHelp(target, h.shell)
		return nil
	}

	// 匹配外部命令分类
	for _, cat := range h.descMgr.getAllCategories() {
		if strings.EqualFold(cat, target) {
			fmt.Printf("📦 分类: %s\n\n", cat)
			if v, ok := h.descMgr.categories.Load(cat); ok {
				cmds := v.([]string)
				for _, name := range cmds {
					if desc, ok := h.descMgr.Get(name); ok {
						fmt.Printf("  %-20s - %s\n", name, desc.Desc)
					}
				}
			}
			return nil
		}
	}

	// 模糊匹配关键字（内置 + 外部）
	matches := []string{}
	for _, cmd := range builtins {
		if strings.Contains(strings.ToLower(cmd.Name()), strings.ToLower(target)) ||
			strings.Contains(strings.ToLower(cmd.Desc()), strings.ToLower(target)) {
			matches = append(matches, cmd.Name())
		}
	}
	h.descMgr.desc.Range(func(k, v any) bool {
		name := k.(string)
		desc := v.(CommandDesc)
		if strings.Contains(strings.ToLower(name), strings.ToLower(target)) ||
			strings.Contains(strings.ToLower(desc.Desc), strings.ToLower(target)) {
			matches = append(matches, name)
		}
		return true
	})

	if len(matches) == 0 {
		fmt.Printf("❌ 未找到与 '%s' 相关的命令\n", target)
		return nil
	}

	fmt.Printf("\n🔍 匹配到 %d 个命令:\n", len(matches))
	for _, m := range matches {
		if cmd, ok := h.shell.commands[m]; ok {
			fmt.Printf("  %-20s - %s\n", cmd.Name(), cmd.Desc())
		} else if desc, ok := h.descMgr.Get(m); ok {
			fmt.Printf("  %-20s - %s\n", m, desc.Desc)
		}
	}
	fmt.Println("\n使用 `help [命令名]` 查看详细帮助")
	return nil
}

// CommandDesc
type CommandDesc struct {
	Category    string   `toml:"-"`
	Desc        string   `toml:"desc"`
	Usage       string   `toml:"usage"`
	Args        []string `toml:"args"`
	Subcommands []string `toml:"subcommands"`
	Flags       []string `toml:"flags"`
	Returns     []string `toml:"returns"`
}

// DescManager
type DescManager struct {
	desc       sync.Map // 命令名 -> CommandDesc
	categories sync.Map // 分类 -> []string
}

func NewDescManager() *DescManager {
	return &DescManager{}
}
func (d *DescManager) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("❌ 解析 desc.toml 失败: %v", err)
	}

	d.desc = sync.Map{}
	d.categories = sync.Map{}

	var walk func(m map[string]interface{}, prefix []string)
	walk = func(m map[string]interface{}, prefix []string) {
		for k, v := range m {
			switch node := v.(type) {
			case map[string]interface{}:
				// 判断是否已经是 CommandDesc 结构
				if _, ok := node["desc"]; ok {
					fullName := append(prefix, k)
					category := fullName[0]
					name := strings.Join(fullName[1:], ".")
					bytes, _ := toml.Marshal(node)
					var desc CommandDesc
					if err := toml.Unmarshal(bytes, &desc); err != nil {
						fmt.Printf("❌ 解析命令 %s 失败: %v\n", strings.Join(fullName, "."), err)
						continue
					}
					desc.Category = category
					d.desc.Store(name, desc)

					val, _ := d.categories.LoadOrStore(category, []string{})
					cmds := append(val.([]string), name)
					d.categories.Store(category, cmds)
				} else {
					// 继续递归
					walk(node, append(prefix, k))
				}
			default:
				// 忽略非 map 节点
			}
		}
	}

	walk(raw, []string{})

	fmt.Printf("✅ desc.toml 已加载，共 %d 条命令，%d 个分类\n", d.countCommands(), len(d.getAllCategories()))
	return nil
}

// Get
func (d *DescManager) Get(name string) (CommandDesc, bool) {
	v, ok := d.desc.Load(name)
	if !ok {
		return CommandDesc{}, false
	}
	return v.(CommandDesc), true
}

// countCommands
func (d *DescManager) countCommands() int {
	cnt := 0
	d.desc.Range(func(_, _ any) bool {
		cnt++
		return true
	})
	return cnt
}

// getAllCategories
func (d *DescManager) getAllCategories() []string {
	keys := []string{}
	d.categories.Range(func(k, _ any) bool {
		keys = append(keys, k.(string))
		return true
	})
	return keys
}

func (d *DescManager) PrintHelp(name string, shell *Shell) {

	// 先检查内置
	if cmd, ok := shell.commands[name]; ok && cmd.IsBuiltin() {
		fmt.Printf("Command: %s\nCategory: %s\n\n", cmd.Name(), cmd.Category())
		fmt.Printf("Usage:\n  %s\n", cmd.Usage())
		if len(cmd.Flags()) > 0 {
			fmt.Println("Flags:")
			for _, v := range cmd.Flags() {
				fmt.Println("   " + v)
			}
		}
		if len(cmd.Subcommands()) > 0 {
			fmt.Println("Subcommands:")
			for _, v := range cmd.Subcommands() {
				fmt.Println("   " + v)
			}
		}
		if len(cmd.Args()) > 0 {
			fmt.Println("Args:")
			for _, v := range cmd.Args() {
				fmt.Println("   " + v)
			}
		}
		if len(cmd.Returns()) > 0 {
			fmt.Println("Returns:")
			for _, v := range cmd.Returns() {
				fmt.Println("   " + v)
			}
		}
		return
	}

	// 外部命令
	c, ok := d.Get(name)
	if !ok {
		fmt.Printf("❌ 未找到命令 %s\n", name)
		return
	}

	fmt.Printf("Command: %s\nCategory: %s\n\n", name, c.Category)
	fmt.Printf("Usage:\n  %s\n", c.Usage)
	if len(c.Flags) > 0 {
		fmt.Println("Flags:")
		for _, v := range c.Flags {
			fmt.Println("   " + v)
		}
	}
	if len(c.Subcommands) > 0 {
		fmt.Println("Subcommands:")
		for _, v := range c.Subcommands {
			fmt.Println("   " + v)
		}
	}
	if len(c.Args) > 0 {
		fmt.Println("Args:")
		for _, v := range c.Args {
			fmt.Println("   " + v)
		}
	}
	if len(c.Returns) > 0 {
		fmt.Println("Returns:")
		for _, v := range c.Returns {
			fmt.Println("   " + v)
		}
	}
}
