package main

import (
	"flyos/modules/routing"
	"flyos/pkg/runtime"
)

func main() {
	rt := runtime.New()

	// 注册模块（未来可通过配置文件或自动扫描）
	rt.RegisterModule(&routing.RouteModule{})
	// rt.RegisterModule(&ids.IDSModule{})

	rt.Start()

	// 模拟执行一条命令（实际应读取用户输入）
	_ = rt.ExecuteCommand("route add", []string{"dst", "192.168.1.0/24", "via", "10.0.0.1"})

	// 阻塞主 goroutine（实际应集成 CLI 循环）
	select {}
}