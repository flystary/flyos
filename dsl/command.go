package dsl

import (
	"fmt"
	"strings"
)

type Command struct {
	Kind    string
	Verb    string
	Subtype string
	Attrs   map[string]interface{}
	Blocks  []Command
}

type ExecutorFunc func(cmd *Command) error

var executors = map[string]ExecutorFunc{}

func Register(kind string, fn ExecutorFunc) {
	executors[strings.ToLower(kind)] = fn
}

func Execute(cmd *Command) error {
	fn, ok := executors[strings.ToLower(cmd.Kind)]
	if !ok {
		return fmt.Errorf("no executor registered for kind '%s'", cmd.Kind)
	}
	return fn(cmd)
}

func ExecuteAll(cmds []Command) error {
	for i := range cmds {
		cmd := &cmds[i]
		// 对 sync，遍历 Blocks
		if cmd.Verb == "sync" && len(cmd.Blocks) > 0 {
			for _, b := range cmd.Blocks {
				fmt.Printf("[%s %s] %s %v\n", cmd.Kind, cmd.Verb, b.Subtype, b.Attrs)
			}
			continue
		}

		fn, ok := executors[strings.ToLower(cmd.Kind)]
		if !ok {
			return fmt.Errorf("no executor registered for kind '%s'", cmd.Kind)
		}
		if err := fn(cmd); err != nil {
			return err
		}
	}
	return nil
}

func PrettyPrint(cmds []Command) {
	for _, c := range cmds {
		fmt.Printf("Kind=%s Verb=%s Subtype=%s Attrs=%v\n", c.Kind, c.Verb, c.Subtype, c.Attrs)
		for _, b := range c.Blocks {
			fmt.Printf(" -> Block Subtype=%s Attrs=%v\n", b.Subtype, b.Attrs)
		}
	}
}
