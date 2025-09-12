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