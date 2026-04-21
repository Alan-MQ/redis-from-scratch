# 阶段2学习任务：RESP + 命令执行器

这一阶段先别急着做持久化、复制或集群。

你现在最需要打通的是：

1. 客户端字节流是怎么被 Redis 解析成 `argv` 的
2. `argv` 又是怎么被路由到具体命令的
3. 命令最终如何读写内存数据库

## 这次我帮你搭好的骨架

- `src/network/resp.go`
  - 放 RESP / inline 命令解析器
  - 目前 inline command 已经可用
  - RESP Array / Bulk String 故意留给你补
- `src/storage/engine.go`
  - 放最小可用内存数据库
  - 已经能存 `RedisObject`
- `src/command/handler.go`
  - 放命令注册和分发
  - `PING` 已经打通
  - `SET/GET/DEL` 故意留给你补
- `src/server/server.go`
  - 服务器已经接到 parser + handler 流程

## 推荐你亲手完成的顺序

### 1. 先补 `readBulkString`

文件：`src/network/resp.go`

原因：
- 这是 RESP 最核心的数据单元
- `SET key value`、`GET key` 的参数本质上都是 bulk string

建议你先想清楚：
- 为什么 bulk string 需要先传长度，再传 payload
- 为什么 Redis 要二进制安全
- 为什么 payload 后面还要再跟一个 `\r\n`

### 2. 再补 `readArrayCommand`

文件：`src/network/resp.go`

原因：
- Redis 客户端发命令时，本质上发的是 `*N` 开头的数组
- 每个数组元素通常又是一个 bulk string

建议你边写边验证：
- `*1\r\n$4\r\nPING\r\n`
- `*2\r\n$4\r\nPING\r\n$5\r\nhello\r\n`
- `*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\nalan\r\n`

### 3. 再补 `SET`

文件：`src/command/handler.go`

原因：
- 这是“写路径”的最小闭环
- 你会第一次把命令参数转成底层对象并存进数据库

思考点：
- Redis 为什么不是直接存裸字符串，而是包一层对象
- `SET` 返回的为什么是 `+OK`

### 4. 再补 `GET`

文件：`src/command/handler.go`

原因：
- 这是“读路径”的最小闭环
- 你会接触类型断言和空值的 RESP 表达

思考点：
- key 不存在时为什么返回 `$-1`
- 为什么 Redis 区分 `nil` 和空字符串

### 5. 最后补 `DEL`

文件：`src/command/handler.go`

原因：
- 你会顺手理解 Redis 为什么很多命令返回 integer reply

思考点：
- `DEL a b c` 为什么返回的是删除成功的数量，而不是简单的 OK/ERR

## 你可以用这些测试作为路标

- `go test ./src/network -v`
- `go test ./src/command -v`
- `go test ./src/storage -v`

如果你把 RESP 补完了，再试：

- `go run main.go`
- `redis-cli -p 6379 ping`
- `redis-cli -p 6379 set name alan`
- `redis-cli -p 6379 get name`

## 我建议你写代码时重点盯住的几个问题

- 协议边界：什么时候该报协议错误
- 空值语义：空字符串、nil、错误响应分别长什么样
- 数据流向：socket -> parser -> argv -> handler -> storage -> RESP response
- 类型系统：为什么 RedisObject 和底层 RedisValue 要分层
