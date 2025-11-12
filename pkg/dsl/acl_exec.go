package dsl

import (
	"fmt"
	"strings"
)

func init() {
	Register("acl", execACL)
}

func execACL(cmd *Command) error {
	switch strings.ToLower(cmd.Verb) {
	case "add", "set", "delete":
		fmt.Printf("[acl %s] subtype=%s attrs=%v\n", cmd.Verb, cmd.Subtype, cmd.Attrs)
	case "sync":
		for _, b := range cmd.Blocks {
			fmt.Printf("[acl sync] subtype=%s attrs=%v\n", b.Subtype, b.Attrs)
		}
	default:
		fmt.Println("[acl] unknown verb", cmd.Verb)
	}
	return nil
}
