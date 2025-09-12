# Redis From Scratch

🚀 一个从零开始构建的功能完整的Redis实现，用于深入学习Redis内部机制和分布式系统概念。

## ✨ 项目特色

- **完整实现**: 涵盖数据结构、网络协议、持久化、集群等核心功能
- **学习导向**: 每个功能都有详细的设计思路和实现指导
- **Go语言**: 使用Go的优雅并发模型，代码简洁易读
- **循序渐进**: 分9个阶段，从基础到高级逐步深入

## 🏗️ 项目结构

```
redis-from-scratch/
├── main.go              # 主程序入口
├── go.mod              # Go模块配置
├── src/                # 源码目录
│   ├── server/         # 服务器核心
│   ├── core/          # 数据结构和命令
│   ├── network/       # 网络协议
│   ├── storage/       # 持久化
│   └── cluster/       # 分布式功能
├── doc/               # 学习文档
│   ├── roadmap.md     # 学习路线图
│   ├── learning_notes.md  # 学习笔记
│   └── conversations.md   # 对话记录
└── .claude/           # AI助手上下文
```

## 🚀 快速开始

### 1. 环境要求
- Go 1.21+
- Git

### 2. 运行项目
```bash
# 克隆项目
git clone <your-repo-url>
cd redis-from-scratch

# 运行服务器
go run main.go

# 或者构建后运行
go build -o redis-server
./redis-server
```

### 3. 测试连接
```bash
# 使用telnet测试基本PING命令
telnet 127.0.0.1 6379
PING
# 应该返回: +PONG
```

### 4. 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定模块测试  
go test ./src/server -v
```

## 📚 学习路径

详细的学习计划请查看 [学习路线图](doc/roadmap.md)

**9个学习阶段**:
1. **项目初始化** - Go环境和基础框架 ✅
2. **基础数据结构** - String、List、Set、Hash、ZSet
3. **命令解析器** - RESP协议和命令执行
4. **网络层** - 并发连接处理
5. **持久化** - RDB快照和AOF日志  
6. **内存管理** - 过期策略和内存优化
7. **主从复制** - 数据同步机制
8. **哨兵系统** - 故障检测和自动切换
9. **集群模式** - 分布式架构

## 🎯 当前进度

- [x] 项目框架搭建
- [x] 基础服务器实现 
- [x] PING/PONG基本功能
- [ ] RESP协议完整实现
- [ ] 核心数据结构实现

## 📖 文档

- [学习路线图](doc/roadmap.md) - 详细的9阶段学习计划
- [学习笔记](doc/learning_notes.md) - 概念理解和学习记录
- [对话记录](doc/conversations.md) - 学习过程中的讨论

## 🤝 参与学习

这是一个学习项目，欢迎：
- 提出问题和改进建议
- 分享学习心得和笔记
- 贡献代码优化和最佳实践

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件