package server

import (
	"net"
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
	
	// 启动服务器
	go func() {
		err := server.Start()
		assert.NoError(t, err)
	}()
	
	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)
	
	// 测试连接
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	assert.NoError(t, err)
	if conn != nil {
		conn.Close()
	}
	
	// 停止服务器
	server.Shutdown()
}

func TestPingPong(t *testing.T) {
	config := &Config{
		Host: "127.0.0.1",  
		Port: 0,
	}
	
	server := New(config)
	
	go func() {
		server.Start()
	}()
	
	time.Sleep(100 * time.Millisecond)
	
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
}

// TODO: Alan 添加更多测试用例
// - 测试多客户端连接
// - 测试错误命令处理
// - 测试连接异常处理