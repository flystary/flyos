// dsl_test.go
package dsl

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// captureOutput 临时替换 stdout 并返回捕获的内容
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestDSLFull(t *testing.T) {
	src := `
# 路由操作
route add static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
route set bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100, 65002:200 ] }
route delete ospf { prefix 192.168.10.0/24 }

# ACL 操作
acl add inbound { src 10.0.0.0/8; dst any; action allow; priority 100 }
acl set outbound { src any; dst 192.168.0.0/16; action deny; log true }

# 声明式同步（最终态）
routes sync {
  static { prefix 10.10.0.0/24; via 10.0.0.1 }
  bgp    { prefix 20.20.0.0/24; nexthop 10.10.10.1; local_pref 300 }
}
acls sync {
  inbound  { src 10.0.0.0/8; action allow }
  outbound { dst 0.0.0.0/0; action deny }
}
`

	t.Run("Parse", func(t *testing.T) {
		p := NewParser(src)
		cmds, err := p.Parse()
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if len(cmds) != 7 {
			t.Errorf("Expected 7 commands, got %d", len(cmds))
		}

		// 验证第一个 route add
		cmd0 := cmds[0]
		if cmd0.Kind != "route" || cmd0.Verb != "add" || cmd0.Subtype != "static" {
			t.Errorf("cmd0 mismatch: %+v", cmd0)
		}
		if prefix, ok := cmd0.Attrs["prefix"].(string); !ok || prefix != "10.0.0.0/24" {
			t.Errorf("cmd0 prefix wrong: %v", cmd0.Attrs["prefix"])
		}
		if track, ok := cmd0.Attrs["track"].(bool); !ok || !track {
			t.Errorf("cmd0 track should be true, got %v", cmd0.Attrs["track"])
		}

		// 验证 BGP community 列表
		cmd1 := cmds[1]
		if comm, ok := cmd1.Attrs["community"].([]string); !ok || len(comm) != 2 {
			t.Errorf("BGP community not parsed as []string: %v", cmd1.Attrs["community"])
		} else if comm[0] != "65001:100" || comm[1] != "65002:200" {
			t.Errorf("Unexpected community: %v", comm)
		}

		// 验证 sync 块
		syncCmd := cmds[5] // routes sync
		if syncCmd.Kind != "route" || syncCmd.Verb != "sync" || len(syncCmd.Blocks) != 2 {
			t.Errorf("Sync command malformed: %+v", syncCmd)
		}
		if syncCmd.Blocks[0].Subtype != "static" || syncCmd.Blocks[1].Subtype != "bgp" {
			t.Errorf("Sync subtypes wrong: %v", syncCmd.Blocks)
		}
	})

	t.Run("Execute", func(t *testing.T) {
		p := NewParser(src)
		cmds, err := p.Parse()
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		output := captureOutput(func() {
			if err := ExecuteAll(cmds); err != nil {
				fmt.Printf("execute error: %v\n", err)
			}
		})

		lines := strings.Split(strings.TrimSpace(output), "\n")
		// 统计所有命令 + sync block，总共 9 行
		if len(lines) != 9 {
			t.Errorf("Expected 9 output lines, got %d:\n%s", len(lines), output)
		}

		// 检查关键输出是否存在
		hasRouteAdd := strings.Contains(output, "[route add] static")
		hasACLAdd := strings.Contains(output, "[acl add] subtype=inbound")
		hasSyncRoute := strings.Contains(output, "[route sync] bgp")
		hasSyncACL := strings.Contains(output, "[acl sync] outbound")

		if !hasRouteAdd || !hasACLAdd || !hasSyncRoute || !hasSyncACL {
			t.Errorf("Missing expected output:\n%s", output)
		}
	})

	t.Run("UnknownKind", func(t *testing.T) {
		badSrc := `firewall add rule { action drop }`
		p := NewParser(badSrc)
		cmds, _ := p.Parse()
		err := ExecuteAll(cmds)
		if err == nil || !strings.Contains(err.Error(), "no executor registered for kind 'firewall'") {
			t.Errorf("Expected unknown kind error, got: %v", err)
		}
	})
}
