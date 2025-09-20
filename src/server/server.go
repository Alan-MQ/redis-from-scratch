package server

import (
	"fmt"
	"net"
	"sync"
)

type Config struct {
	Host string
	Port int
}

type Server struct {
	config   *Config
	listener net.Listener
	clients  map[string]*Client
	mutex    sync.RWMutex
	shutdown chan struct{}
	// storage        *storage.Engine
	// commandHandler *command.Handler

	// TODO: Alan 添加数据存储引擎
	// storage *storage.Engine

	// TODO: Alan 添加命令处理器
	// commandHandler *command.Handler
}

type Client struct {
	conn net.Conn
	id   string
	// TODO: Alan 添加客户端状态管理
}

func New(config *Config) *Server {
	return &Server{
		config:   config,
		clients:  make(map[string]*Client),
		shutdown: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	fmt.Printf("📡 Redis server listening on %s\n", addr)

	// TODO: Alan 实现连接处理循环
	for {
		select {
		case <-s.shutdown:
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				// 检查是否是因为服务器关闭导致的错误
				select {
				case <-s.shutdown:
					return nil
				default:
					fmt.Printf("❌ Accept error: %v\n", err)
					continue
				}
			}

			// 处理新连接
			go s.handleClient(conn)
		}
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()

	clientID := conn.RemoteAddr().String()
	client := &Client{
		conn: conn,
		id:   clientID,
	}

	// 注册客户端
	s.mutex.Lock()
	s.clients[clientID] = client
	s.mutex.Unlock()

	fmt.Printf("✅ Client connected: %s\n", clientID)

	// TODO: Alan 实现RESP协议解析和命令处理
	// 这里现在只是一个占位符，返回简单的PONG响应
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}

		// 		RESP 用不同的首字母标记不同类型，后面跟 \r\n（CRLF）结束。常见类型（RESP2）：
		// Simple String（简单字符串）
		// 前缀 +，直到 CRLF 的文本（通常用于状态回复）。例：+OK\r\n
		// Error（错误）
		// 前缀 -，直到 CRLF 的错误文本。例：-ERR unknown command 'FOO'\r\n
		// Integer（整数）
		// 前缀 :，直到 CRLF 的十进制整数。例：:1000\r\n
		// Bulk String（批量字符串，二进制安全）
		// 前缀 $ 接长度 \r\n，然后是精确的 N 字节数据，再跟一个 \r\n。
		// 例：$6\r\nfoobar\r\n（数据是 6 字节 foobar）
		// 空字符串：$0\r\n\r\n
		// Null（nil）：$-1\r\n（表示没有值）
		// Array（数组 / 多批量）
		// 前缀 * 接元素个数 \r\n，然后按顺序写每个元素（元素可以是任意 RESP 类型，支持嵌套）。
		// 例：*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
		// Null array：*-1\r\n
		command := string(buffer[:n])
		fmt.Printf("📨 Received: %s from %s\n", command, clientID)

		// 简单的PING-PONG响应
		if command == "PING\r\n" {
			conn.Write([]byte("+PONG\r\n"))
		} else {
			conn.Write([]byte("-ERR unknown command\r\n"))
		}
	}

	// 清理客户端连接
	s.mutex.Lock()
	delete(s.clients, clientID)
	s.mutex.Unlock()

	fmt.Printf("❌ Client disconnected: %s\n", clientID)
}

func (s *Server) Shutdown() {
	close(s.shutdown)

	if s.listener != nil {
		s.listener.Close()
	}

	// 关闭所有客户端连接
	s.mutex.Lock()
	for _, client := range s.clients {
		client.conn.Close()
	}
	s.mutex.Unlock()
}
