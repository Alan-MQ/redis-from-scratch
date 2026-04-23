# 阶段2扩展任务：Hash 命令接线

这一轮的目标是把已经做好的 `Dict` 真正变成 Redis 里的 Hash 类型。

你现在已经有了：

- `Dict` 的基础能力
- `RedisObject` 和类型系统
- RESP 解析和命令分发
- String / List 的命令接线经验

所以现在做 Hash 是最自然的一步。

## 这一步你会重点学到什么

1. 一个“通用底层结构”如何被包装成特定 Redis 类型
2. `HSET/HGET/HDEL/HEXISTS/HGETALL/HLEN` 的返回值语义
3. 为什么 Redis 很多命令不是简单返回 `OK`
4. `WRONGTYPE` 如何贯穿所有命令族

## 我这次帮你搭好的骨架

- `src/core/hash.go`
  - 新增 `RedisHash`
  - 基础 field/value 封装已接好
- `src/core/hash_test.go`
  - 验证 `RedisHash` 的基础行为
- `src/command/hash_commands_test.go`
  - 给你准备好了命令层的测试路标
- `src/command/handler.go`
  - 下一步会注册 Hash 命令并接 helper

## 推荐你亲手完成的顺序

### 1. 先实现 `getOrCreateHash` / `getExistingHash`

为什么：
- 这一步会重复你在 List 命令里已经掌握的对象模型模式
- 你会更熟悉“Redis 类型检查 + WRONGTYPE”的统一写法

### 2. 再实现 `HSET`

为什么先做它：
- 这是写路径
- 它最能体现 Hash 和 Dict 的关系

建议重点想清楚：
- 为什么 `HSET` 返回的是“新增字段数量”
- 为什么 `HSET key field value` 和 `HSET key f1 v1 f2 v2` 都合法

### 3. 再实现 `HGET` / `HEXISTS` / `HLEN`

这三个是最好的读路径练习：
- `HGET` 练 null bulk string
- `HEXISTS` 练 integer reply
- `HLEN` 练缺失 key 的零值语义

### 4. 最后实现 `HGETALL` / `HDEL`

原因：
- `HGETALL` 会同时碰到数组响应和结果展开
- `HDEL` 会让你再次体会 Redis 为什么爱返回 integer reply

## 你可以这样验证

- `go test ./src/core ./src/command -v`

实现完以后再手动试：

- `HSET profile name alan city hangzhou`
- `HGET profile name`
- `HEXISTS profile city`
- `HLEN profile`
- `HGETALL profile`
- `HDEL profile city`

## 一个很重要的建模点

Hash 和 Set 都可以基于 `Dict`，但它们的语义不同：

- Set 只关心 member 是否存在
- Hash 关心 `field -> value`

这正是 Redis 很有代表性的设计习惯：
**底层复用结构，上层暴露不同命令语义。**
