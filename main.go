// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
)

// env merge
var baseEnv []string

func init() {
	baseEnv = os.Environ()
}

// mergeEnv 合并系统环境变量与自定义配置环境变量。
// 支持 PATH 数组合并、$PATH 占位符、DEBUG 打印。
func mergeEnv(custom map[string]interface{}) []string {
	envMap := map[string]string{}

	// 1️⃣ 导入系统环境
	for _, kv := range baseEnv {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	sysPath := envMap["PATH"]
	sep := string(os.PathListSeparator)

	// 2️⃣ 处理自定义环境变量
	for k, v := range custom {
		switch val := v.(type) {
		case string:
			// 支持 $PATH 展开
			if k == "PATH" {
				val = strings.ReplaceAll(val, "$PATH", sysPath)
			}
			envMap[k] = val

		case []interface{}:
			// 支持 PATH 数组合并
			if k == "PATH" {
				paths := []string{}
				for _, p := range val {
					if s, ok := p.(string); ok {
						if s == "$PATH" {
							paths = append(paths, strings.Split(sysPath, sep)...)
						} else if strings.Contains(s, "$PATH") {
							paths = append(paths, strings.Split(strings.ReplaceAll(s, "$PATH", sysPath), sep)...)
						} else {
							paths = append(paths, s)
						}
					}
				}
				envMap[k] = strings.Join(paths, sep)
			} else {
				// 非 PATH 数组转为逗号字符串
				strVals := []string{}
				for _, p := range val {
					strVals = append(strVals, fmt.Sprintf("%v", p))
				}
				envMap[k] = strings.Join(strVals, ",")
			}

		default:
			envMap[k] = fmt.Sprintf("%v", val)
		}
	}

	// 3️⃣ DEBUG 模式打印
	debug := strings.EqualFold(envMap["DEBUG"], "true")
	if debug {
		fmt.Println("🪴 [flyos] Merged environment variables:")
		for k, v := range envMap {
			if k == "PATH" {
				fmt.Printf("  %-10s = %s\n", k, v)
			} else {
				fmt.Printf("  %-10s = %q\n", k, v)
			}
		}
	}

	// 4️⃣ 转换为 os.Environ 格式
	result := []string{}
	for k, v := range envMap {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// 配置结构
type Config struct {
	CommandsDirs []string               `toml:"commands_dirs"`
	Excludes     []string               `toml:"excludes"`
	Env          map[string]interface{} `toml:"env"`
}

// Config
func parseConfig() (*Config, error) {
	data, err := os.ReadFile(".config.toml")
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	envStr := make(map[string]interface{})
	sysPath := os.Getenv("PATH")
	sep := string(os.PathListSeparator)

	for k, v := range cfg.Env {
		switch val := v.(type) {
		case string:
			// 单字符串 PATH 支持 $PATH 展开
			if k == "PATH" {
				envStr[k] = strings.ReplaceAll(val, "$PATH", sysPath)
			} else {
				envStr[k] = val
			}

		case []interface{}:
			paths := []string{}
			for _, p := range val {
				if s, ok := p.(string); ok {
					// 展开 $PATH
					if s == "$PATH" {
						paths = append(paths, strings.Split(sysPath, sep)...)
					} else if strings.Contains(s, "$PATH") {
						paths = append(paths, strings.Split(strings.ReplaceAll(s, "$PATH", sysPath), sep)...)
					} else {
						paths = append(paths, s)
					}
				}
			}
			envStr[k] = strings.Join(paths, sep)

		default:
			envStr[k] = fmt.Sprintf("%v", val)
		}
	}

	fmt.Println("🌍 已加载环境变量:")
	for k, v := range envStr {
		fmt.Printf("  %s=%s\n", k, v)
	}

	return &Config{
		CommandsDirs: cfg.CommandsDirs,
		Excludes:     cfg.Excludes,
		Env:          envStr,
	}, nil
}

// REPL
type REPL struct {
	shell *Shell
	desc  *DescManager
	rl    *readline.Instance
}

func NewREPL(shell *Shell, desc *DescManager) (*REPL, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:      "flyos> ",
		HistoryFile: "/tmp/flyos_history",
	})
	if err != nil {
		return nil, err
	}
	return &REPL{shell: shell, desc: desc, rl: l}, nil
}

func (r *REPL) Loop() {
	defer r.rl.Close()
	for {
		line, err := r.rl.Readline()
		if err != nil {
			break
		}
		args := strings.Fields(strings.TrimSpace(line))
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "exit":
			fmt.Println("👋 Bye!")
			return
		case "list":
			r.shell.List()
		default:
			r.shell.RunCommand(args)
		}
	}
}

// Main
func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Printf("❌ 启动失败: %v\n", err)
		return
	}

	shell := NewShell(cfg.Env)
	shell.env["USER"] = "fly"
	shell.env["VERSION"] = "1.0.0"
	// 内置命令注册
	shell.Register(&EnvCommand{})
	shell.Register(&ExitCommand{})
	shell.Register(&ListCommand{})

	desc := NewDescManager()
	_ = desc.Load("desc.toml")

	// 注册 HelpCommand（关键）
	helpCmd := NewHelpCommand(desc, shell)
	shell.Register(helpCmd)
	shell.LoadCommands(cfg, desc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// fsnotify 监听
	go func() {
		watcher, _ := fsnotify.NewWatcher()
		defer watcher.Close()
		_ = watcher.Add(".")
		var debounce *time.Timer
		for {
			select {
			case ev := <-watcher.Events:
				switch filepath.Base(ev.Name) {
				case ".config.toml":
					if ev.Op&fsnotify.Write != 0 {
						if debounce != nil {
							debounce.Stop()
						}
						debounce = time.AfterFunc(300*time.Millisecond, func() {
							cfg, err := parseConfig()
							if err != nil {
								fmt.Println("❌ reload config failed:", err)
								return
							}
							shell.LoadCommands(cfg, desc)
						})
					}
				case "desc.toml":
					if ev.Op&fsnotify.Write != 0 {
						if debounce != nil {
							debounce.Stop()
						}
						debounce = time.AfterFunc(300*time.Millisecond, func() {
							_ = desc.Load("desc.toml")
						})
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	repl, err := NewREPL(shell, desc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("🚀 flyos REPL 启动！输入 'help' 查看命令 'exit' 退出环境 ")
	repl.Loop()
}
