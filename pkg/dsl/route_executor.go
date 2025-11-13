package dsl

import (
	"fmt"
	"strings"
)

func init() {
	Register("route", execRoute)
}

func execRoute(cmd *Command) error {
	switch strings.ToLower(cmd.Verb) {
	case "add":
		fmt.Println("[route add]", cmd.Subtype, cmd.Attrs)
	case "set":
		fmt.Println("[route set]", cmd.Subtype, cmd.Attrs)
	case "delete":
		fmt.Println("[route del]", cmd.Subtype, cmd.Attrs)
	case "sync":
		for _, b := range cmd.Blocks {
			fmt.Println("[route sync]", b.Subtype, b.Attrs)
		}
	default:
		fmt.Println("[route] unknown verb", cmd.Verb)
	}
	return nil
}
