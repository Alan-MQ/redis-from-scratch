package server

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerStartStop(t *testing.T) {
	config := &Config{
		Host: "127.0.0.1",
		Port: 0, // 让系统自动分配端口
	}

	server := New(config)

	errCh := make(chan error, 1)

	// 启动服务器
	go func() {
		errCh <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	if server.listener == nil {
		select {
		case err := <-errCh:
			if err != nil && strings.Contains(err.Error(), "operation not permitted") {
				t.Skip("sandbox 环境不允许绑定本地 TCP 端口，跳过 server 集成测试")
				return
			}
			assert.NoError(t, err)
		default:
			t.Fatal("server listener was not initialized")
		}
	}

	// 测试连接
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	assert.NoError(t, err)
	if conn != nil {
		conn.Close()
	}

	// 停止服务器
	server.Shutdown()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestPingPong(t *testing.T) {
	config := &Config{
		Host: "127.0.0.1",
		Port: 0,
	}

	server := New(config)

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	if server.listener == nil {
		select {
		case err := <-errCh:
			if err != nil && strings.Contains(err.Error(), "operation not permitted") {
				t.Skip("sandbox 环境不允许绑定本地 TCP 端口，跳过 server 集成测试")
				return
			}
			assert.NoError(t, err)
		default:
			t.Fatal("server listener was not initialized")
		}
	}

	// 连接并发送PING
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	assert.NoError(t, err)
	defer conn.Close()

	// 发送PING命令
	_, err = conn.Write([]byte("PING\r\n"))
	assert.NoError(t, err)

	// 读取响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	assert.NoError(t, err)

	response := string(buffer[:n])
	assert.Equal(t, "+PONG\r\n", response)

	server.Shutdown()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("server did not stop in time")
	}
}

// TODO: Alan 添加更多测试用例
// - 测试多客户端连接
// - 测试错误命令处理
// - 测试连接异常处理
