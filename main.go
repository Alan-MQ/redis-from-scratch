package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"redis-from-scratch/src/server"
)

func main() {
	fmt.Println("🚀 Redis From Scratch - Starting...")
	
	// TODO: Alan 读取配置文件，设置服务器参数
	config := &server.Config{
		Host: "127.0.0.1",
		Port: 6379,
	}
	
	// 创建Redis服务器实例
	redisServer := server.New(config)
	
	// 启动服务器
	go func() {
		if err := redisServer.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	fmt.Println("🛑 Shutting down Redis server...")
	redisServer.Shutdown()
	fmt.Println("✅ Server stopped")
}