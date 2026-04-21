package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"redis-from-scratch/src/command"
	"redis-from-scratch/src/network"
	"redis-from-scratch/src/storage"
)

type Config struct {
	Host string
	Port int
}

type Server struct {
	config         *Config
	listener       net.Listener
	clients        map[string]*Client
	mutex          sync.RWMutex
	shutdown       chan struct{}
	storage        *storage.Engine
	commandHandler *command.Handler
}

type Client struct {
	conn net.Conn
	id   string
	// TODO: Alan 添加客户端状态管理
}

func New(config *Config) *Server {
	engine := storage.NewEngine()

	return &Server{
		config:         config,
		clients:        make(map[string]*Client),
		shutdown:       make(chan struct{}),
		storage:        engine,
		commandHandler: command.NewHandler(engine),
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

	parser := network.NewParser(conn)
	for {
		args, err := parser.ReadCommand()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			fmt.Printf("❌ Parse error from %s: %v\n", clientID, err)
			if _, writeErr := conn.Write(command.ErrorResult(err.Error()).Encode()); writeErr != nil {
				break
			}

			// 解析失败时，当前连接里的字节流状态可能已经不一致，直接断开最安全。
			break
		}

		fmt.Printf("📨 Received args: %v from %s\n", args, clientID)

		result, execErr := s.commandHandler.Execute(args)
		if _, err := conn.Write(result.Encode()); err != nil {
			break
		}

		if execErr != nil {
			fmt.Printf("🧠 Learning TODO for %s: %v\n", clientID, execErr)
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
