# 阶段3预热任务：EXPIRE / TTL

这一轮是第一次明显进入“这很 Redis”的区域。

因为从这里开始，你写的不再只是“把值存进去再读出来”，而是开始处理：

- key 的生命周期
- 元数据如何影响命令结果
- 为什么 Redis 的一些命令会返回 `-2 / -1 / 正整数`

## 为什么现在做这一步最合适

你已经有了：

- `RedisObject.ExpireTime`
- 存储层的被动过期入口（`Get` 时会检查过期）
- 基础命令执行通路

所以现在补 `EXPIRE/TTL`，是把“对象元数据”真正变成命令语义的最好时机。

## 你会重点学到什么

1. `EXPIRE key seconds` 为什么返回 1/0
2. `TTL key` 为什么有 3 类结果：
   - `-2`：key 不存在
   - `-1`：key 存在，但没有过期时间
   - `>= 0`：剩余秒数
3. Redis 的“被动过期”是怎么工作的
4. 对象模型为什么比裸 `map[string]string` 更重要

## 这次我先帮你搭好的骨架

- `src/command/expire_commands_test.go`
  - 把最关键的行为路标先写好了
  - 包括 `TTL=-1`、`TTL=-2`、有效 TTL、非法参数
- `src/command/handler.go`
  - 下一步会注册 `EXPIRE/TTL`

## 推荐你亲手完成的顺序

### 1. 先实现 `TTL`

为什么：
- 它最能帮你吃透 Redis 的特殊返回值语义
- 写完它你会马上理解 Redis 的 key 生命周期模型

### 2. 再实现 `EXPIRE`

为什么：
- 它是设置生命周期的写路径
- 它和 `TTL` 组合起来，刚好形成一个完整闭环

## 你可以这样验证

- `go test ./src/command -v`

实现完以后再手动试：

- `SET session abc`
- `TTL session`
- `EXPIRE session 10`
- `TTL session`
- `TTL missing`

## 一个很重要的 Redis 心智模型

从这一步开始，你要把 key 想成：

**值 + 类型 + 过期时间 + 访问时间**

也就是说，Redis 真正管理的不是“裸值”，而是“带元数据的对象”。
