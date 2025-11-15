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

# 网络接口操作
# nic
nic set enp1s0 { speed 1000; duplex full; link up }
nic set usb4g { speed 1000; duplex full; link up }
nic set usb5g { speed 1000; duplex full; link up }
nic list
{
    enp1s0 { speed 1000; duplex full; link up }
    enp1s1 { speed 1000; duplex full; link up }
}

# bond操作
bond add bond0 {
    mode lacp;                     // active-backup, balance-rr, lacp(802.3ad)
    members [ enp1s0, enp1s1 ];    // 引用 nic 名称
    miimon 100;
    admin up;
}
bond set bond0 { mode active-backup }
bond delete bond0
bond sync {
    uplink-bond {
        mode lacp;
        members [ enp1s0, enp1s1 ];
        admin up;
    }
}
# spand
spand add mirror-to-ids {
source {
	interfaces [ eth0, bond0 ];
	direction both;   // rx, tx, both
}
destination interface eth3;   // 镜像输出口
truncate 128;                 // 可选：截断字节数
}
spand delete mirror-to-ids
spand sync {
    debug-mirror {
        source { interfaces [ vlan100 ]; direction rx }
        destination interface tap0;
    }
}

#
interface add eth0 {
    admin up;
    ip addr 192.168.1.1/24;
    mtu 1500;
}

interface add loopback {
    type loopback;
    ip addr 10.255.255.1/32;
}
	vlan add vlan100 {
    parent bond0;                // 可基于 bond / eth / bridge
    vid 100;
    admin up;
    ip addr 192.168.100.1/24;
}
# vlan
vlan add bridge br0 {
    type bridge;
    vids [ 100, 200 ];
    members [ eth2, eth3 ];
}
# gre
gre add gre-aws {
	local 203.0.113.10;          // 本地 endpoint
	remote 198.51.100.20;        // 对端 endpoint
	key 1001;                    // 可选
	admin up;
	ttl 255;
}
gre set gre-aws {
	local 203.0.113.10;          // 本地 endpoint
	remote 198.51.100.20;        // 对端 endpoint
	key 1001;                    // 可选
	admin up;
	ttl 255;
}
gre delete gre-aws
gre sync {
	gre-aws {local 203.0.113.10; remote 198.51.100.20; key 1001; ttl 255;}
	gre-sws {local 203.0.113.20; remote 198.51.100.10; key 1002; ttl 255;}
}

ipsec add ipsec-vpc {
    local 203.0.113.10;
    remote 52.10.20.30;
    psk "s3cr3t!";               // 或引用 secret store
    ike_version 2;
    encryption aes256;
    integrity sha256;
    dh_group modp2048;
    lifetime 3600;               // seconds
    admin up;
}

ipsec set ipsec-vpc {
	local 203.0.113.10;
	remote 52.10.20.30;
	psk "s3cr3t!";               // 或引用 secret store
	ike_version 2;
	encryption aes256;
	integrity sha256;
	dh_group modp2048;
	lifetime 3600;               // seconds
	admin up;
}

ipsec del ipsec-vpc
ipsec sync {
	ipsec-aws {
	local 203.0.113.8;
	remote 52.10.20.9;
	psk "s3cr3t!";               // 或引用 secret store
	ike_version 2;
	encryption aes256;
	integrity sha256;
	dh_group modp2048;
	lifetime 3600;               // seconds
	admin up;
	}
	{
	local 203.0.113.40;
	remote 52.10.20.40;
	psk "s3cr3t!";               // 或引用 secret store
	ike_version 2;
	encryption aes256;
	integrity sha256;
	dh_group modp2048;
	lifetime 3600;               // seconds
	admin up;
	}
}

# 路由操作
route add static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
route set static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
route delete static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
route add bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100 ] }
route set bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100, 65002:200 ] }
route add ospf { prefix 192.168.10.0/24; area 0.0.0.0; type external }
route delete ospf { prefix 192.168.10.0/24 }
route set pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth1 }
route delete pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth1 }

# ACL 操作
acl add inbound { src 10.0.0.0/8; dst any; action allow; priority 100 }
acl set outbound { src any; dst 192.168.0.0/16; action deny; log true }

# 声明式同步
route sync {
	static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
	bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100 ] }
	ospf { prefix 192.168.10.0/24; area 0.0.0.0; type external }
	pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth1 }
	static { prefix 20.0.0.0/24; via 192.168.2.1; dev eth1; track yes }
}

acl sync {
  inbound  { src 10.0.0.0/8; action allow }
  outbound { dst 0.0.0.0/0; action deny }
}

route list bgp
{
	bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100 ] }
	bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100 ] }
}

route list ospf
{
	ospf { prefix 192.168.10.0/24; area 0.0.0.0; type external }
	ospf { prefix 192.168.10.0/24; area 0.0.0.0; type external }
}
route list pbr
{
	pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth1 }
	pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth2 }
}
acl list
{
  inbound  { src 10.0.0.0/8; action allow }
  outbound { dst 0.0.0.0/0; action deny }
}
route list
{
	static { prefix 10.0.0.0/24; via 192.168.1.1; dev eth0; track yes }
	bgp { prefix 172.16.0.0/16; local_pref 200; community [ 65001:100 ] }
	ospf { prefix 192.168.10.0/24; area 0.0.0.0; type external }
	pbr { prefix 10.1.0.0/16; fwmark 100; priority 1000; iif eth1 }
	static { prefix 20.0.0.0/24; via 192.168.2.1; dev eth1; track yes }
}
# nat
nat add snat-out {
    type snat;
    match {
        src 10.0.0.0/8;
        out_interface bond0;     // 出接口
    }
    to 203.0.113.10;
}

nat add dnat-tunnel {
    type dnat;
    match {
        in_tunnel ipsec-vpc;     // 关键：来自隧道
        proto tcp;
        port 443;
    }
    to 192.168.10.100:443;
}

nat add masq-vlan {
    type masquerade;
    match { out_interface vlan200 }
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
