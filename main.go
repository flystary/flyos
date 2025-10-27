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

func mergeEnv(custom map[string]string) []string {
	env := make([]string, len(baseEnv))
	copy(env, baseEnv)
	for k, v := range custom {
		env = append(env, k+"="+v)
	}
	return env
}

// 配置结构
type Config struct {
    CommandsDirs []string      			`toml:"commands_dirs"`
    Excludes     []string      			`toml:"excludes"`
    Env          map[string]interface{} `toml:"env"` // 允许值为 string 或 []string
}

func (c *Config) NormalizeEnv() map[string]string {
    result := make(map[string]string)
    
    for key, rawVal := range c.Env {
        switch val := rawVal.(type) {
        case string:
            result[key] = val
        case []interface{}:
            // TOML 解析数组为 []interface{}
            parts := make([]string, 0, len(val))
            for _, v := range val {
                if s, ok := v.(string); ok {
                    parts = append(parts, s)
                }
            }
            result[key] = strings.Join(parts, ":")
        case []string:
            // 某些解析器可能直接返回 []string
            result[key] = strings.Join(val, ":")
        default:
            // 兜底：转为字符串（如数字、bool）
            result[key] = fmt.Sprintf("%v", val)
        }
    }
    return result
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
	return &cfg, nil
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

	envMap := cfg.NormalizeEnv()
	shell := NewShell(envMap)
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
	fmt.Println("🚀 FlyOS REPL 已启动！💡 输入 help 查看命令，输入 exit 安全退出 " )
	repl.Loop()
}
