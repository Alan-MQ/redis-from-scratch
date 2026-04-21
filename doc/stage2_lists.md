# 阶段2扩展任务：List 命令接线

你现在已经打通了：

- RESP 请求解析
- 基础命令分发
- String 类型的 `SET/GET/DEL`

下一步最适合继续的是把你已经实现好的双向链表接到 Redis 命令里。

## 这一步你会学到什么

1. 一个底层数据结构是如何暴露成 Redis 命令语义的
2. 为什么 Redis 要做统一的 `WRONGTYPE` 错误处理
3. RESP 数组响应是怎么编码的
4. 命令行为和底层实现之间，哪里需要做“协议层转换”

## 我这次已经帮你搭好的骨架

- `src/command/result.go`
  - 新增了 `ArrayResult`
  - 这样 `LRANGE` 可以返回 RESP Array
- `src/command/handler.go`
  - 注册了 `LPUSH/RPUSH/LPOP/RPOP/LRANGE`
  - 加了 `getOrCreateList` 和 `getExistingList` 两个 helper
  - 命令主体故意留给你实现
- `src/command/list_commands_test.go`
  - 给你准备好了列表命令的测试路标

## 推荐你亲手完成的顺序

### 1. 先实现 `LPUSH`

文件：`src/command/handler.go`

为什么先写它：
- 这是第一次把“已有复杂数据结构”挂进数据库
- 你会看到 Redis 的“key 不存在就创建”语义

思考点：
- 为什么 `LPUSH key a b` 最终列表顺序是 `b, a`
- 为什么 `LPUSH` 返回的是长度，而不是 `OK`

### 2. 再实现 `RPUSH`

文件：`src/command/handler.go`

为什么紧跟它：
- 逻辑和 `LPUSH` 几乎对称
- 你可以借这个机会刻意比较“头插”和“尾插”的语义差别

### 3. 再实现 `LPOP` / `RPOP`

文件：`src/command/handler.go`

思考点：
- key 不存在时为什么返回 nil bulk string
- 空列表和不存在的 key 在协议层上应该怎么表现

### 4. 最后实现 `LRANGE`

文件：`src/command/handler.go`

这是这一轮最值得你啃的点。

因为它会同时碰到：
- 字符串参数转整数
- 负数索引
- RESP Array 编码
- 列表遍历结果到协议响应的转换

## 你可以这样验证

- `go test ./src/command -v`

实现完以后再手动试：

- `LPUSH tasks b a`
- `RPUSH tasks c`
- `LRANGE tasks 0 -1`
- `LPOP tasks`
- `RPOP tasks`

## 一个很重要的 Redis 语义

`GET` 只能读 string。

同理：
- `LPUSH/RPUSH/LPOP/RPOP/LRANGE` 只能操作 list
- 如果 key 存的是别的类型，应该返回：

`WRONGTYPE Operation against a key holding the wrong kind of value`

这就是 Redis 命令层和对象模型结合得非常紧的地方。
