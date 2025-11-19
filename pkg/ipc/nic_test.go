package ipc_test

import (
	"flyos/pkg/ipc"
	"os"
	"testing"
	"time"
)

func TestIPC_NIC(t *testing.T) {
	sock := "/tmp/test_ipc.sock"
	_ = os.Remove(sock)

	// 启动服务端
	srv := ipc.NewServer(sock)
	nicCfg := &ipc.NICConfig{}
	ipc.RegisterModule(nicCfg)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	// 等待服务端启动
	time.Sleep(5 * time.Millisecond)

	select {
	case err := <-errCh:
		t.Fatal("Server start error:", err)
	default:
		// no error, continue
	}

	cli := ipc.NewClient(sock)

	// Helper: 发送请求并打印
	send := func(method string, payload any) {
		resp, err := cli.SendTyped(method, payload)
		if err != nil {
			t.Fatalf("SendTyped %s error: %v", method, err)
		}
		t.Logf("Response for %s: %+v", method, resp)
	}

	//  测试 NIC 功能
	cfg := &ipc.NICConfig{
		Name: "eth0",
		Up:   true,
		IPs:  []string{"192.168.1.100/24"},
	}

	// set
	send("nic.set", cfg)

	// up
	send("nic.up", cfg)

	// addip
	send("nic.addip", &ipc.NICConfig{
		Name: "eth0",
		IPs:  []string{"192.168.1.101/24"},
	})

	// down
	send("nic.down", cfg)
}
