package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// Shell
type Shell struct {
	mu       sync.RWMutex
	commands map[string]Command
	env      map[string]interface{}
}

func NewShell(customEnv map[string]interface{}) *Shell {
	env := make(map[string]interface{})
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	for k, v := range customEnv {
		env[k] = v
	}
	return &Shell{
		commands: make(map[string]Command),
		env:      env,
	}
}

func (s *Shell) Register(cmd Command) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.commands[cmd.Name()] = cmd
}

func (s *Shell) List() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 分区
	builtinCategories := make(map[string][]Command)
	externalCategories := make(map[string][]Command)

	for _, cmd := range s.commands {
		if cmd.IsBuiltin() {
			builtinCategories[cmd.Category()] = append(builtinCategories[cmd.Category()], cmd)
		} else {
			externalCategories[cmd.Category()] = append(externalCategories[cmd.Category()], cmd)
		}
	}

	// 内置命令输出
	fmt.Println("📦 内置命令:")
	if len(builtinCategories) == 0 {
		fmt.Println("  <无>")
	} else {
		bcats := make([]string, 0, len(builtinCategories))
		for c := range builtinCategories {
			bcats = append(bcats, c)
		}
		sort.Strings(bcats)
		for _, cat := range bcats {
			fmt.Println("📂 分类:")
			fmt.Printf("\n[%s]\n", cat)
			cmds := builtinCategories[cat]
			sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name() < cmds[j].Name() })
			for _, c := range cmds {
				desc := c.Desc()
				if desc == "" {
					desc = "<暂无描述>"
				}
				fmt.Printf("  %-10s - %s\n", c.Name(), desc)
			}
		}
	}

	// 外部命令输出
	fmt.Println("\n📦 外部命令:")
	if len(externalCategories) == 0 {
		fmt.Println("  <无>")
	} else {
		ecats := make([]string, 0, len(externalCategories))
		for c := range externalCategories {
			ecats = append(ecats, c)
		}
		sort.Strings(ecats)
		fmt.Println("📂 分类:")
		for _, cat := range ecats {
			fmt.Printf("\n[%s]\n", cat)
			cmds := externalCategories[cat]
			sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name() < cmds[j].Name() })
			for _, c := range cmds {
				desc := c.Desc()
				if desc == "" {
					desc = "<暂无描述>"
				}
				fmt.Printf("  %-10s → %-20s %s\n", c.Name(), c.Path(), desc)
			}
		}
	}
}

// RunCommand
func (s *Shell) RunCommand(args []string) {
	if len(args) == 0 {
		return
	}
	s.mu.RLock()
	cmd, ok := s.commands[args[0]]
	s.mu.RUnlock()
	if !ok {
		fmt.Printf("❌ 未找到命令: %s\n", args[0])
		return
	}
	if err := cmd.Execute(args, s.env); err != nil {
		fmt.Printf("💥 执行失败 [%s]: %v\n", args[0], err)
	}
}

// FuzzyFind 支持命令模糊匹配
func (s *Shell) FuzzyFind(keyword string) []Command {
	s.mu.RLock()
	defer s.mu.RUnlock()
	results := []Command{}
	kw := strings.ToLower(keyword)
	for name, cmd := range s.commands {
		if strings.Contains(strings.ToLower(name), kw) ||
			strings.Contains(strings.ToLower(cmd.Desc()), kw) {
			results = append(results, cmd)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Name() < results[j].Name() })
	return results
}

// LoadCommands
func (s *Shell) LoadCommands(cfg *Config, descMgr *DescManager) {
	excluded := make(map[string]bool)
	for _, e := range cfg.Excludes {
		excluded[e] = true
	}
	newMap := make(map[string]Command)
	for _, dir := range cfg.CommandsDirs {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if d.IsDir() || excluded[d.Name()] {
				return nil
			}
			if isExecutable(path) {
				name := d.Name()
				category := "default"
				// 如果 descMgr 有记录，则取分类
				if desc, ok := descMgr.Get(name); ok {
					category = desc.Category
				}
				newMap[name] = &FileCommand{
					name:     name,
					path:     path,
					category: category,
				}
			}
			return nil
		})
	}

	s.mu.Lock()
	for k, v := range newMap {
		s.commands[k] = v
	}
	s.mu.Unlock()
	fmt.Printf("✅ 已加载 %d 个外部命令\n", len(newMap))
}

// 文件扫描
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	if runtime.GOOS != "windows" {
		if info.Mode().Perm()&0111 != 0 {
			return true
		}
		f, err := os.Open(path)
		if err != nil {
			return false
		}
		defer f.Close()
		buf := make([]byte, 2)
		n, _ := f.Read(buf)
		return n == 2 && buf[0] == '#' && buf[1] == '!'
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".exe", ".bat", ".cmd", ".com", ".ps1", ".vbs", ".js", ".sh", ".py", ".pl":
		return true
	default:
		return false
	}
}
