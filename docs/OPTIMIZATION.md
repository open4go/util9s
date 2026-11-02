# util9s 优化方案

| # | 优先级 | 问题 | 位置 | 建议 | 工作量 |
|---|---|---|---|---|---|
| 1 | 高 | 公开仓库里硬编码了旧 AES 密钥 `LegacyHardcodedKey`，`Decrypt` 在主密钥解密失败后会用它再试一次，等于所有用旧密钥加密的数据都可以被任何人解开 | `crypto/aes.go:20` | 写一个迁移脚本，用主密钥重新加密存量数据，然后删除旧密钥的回退逻辑；需要兼容的话，从环境变量读取旧密钥 | M |
| 2 | 高 | `LoadPrivateKey` 失败时把**私钥原文**写进日志（`WithField("key", key)`） | `private.go:39` | 删掉这个字段，最多记录长度或指纹 | S |
| 3 | 高 | HTTP 用的是 `http.Post`/`http.Get`（默认客户端没有超时），第三方接口卡住会拖住 goroutine；`Post` 不检查状态码 | `http.go` | 用共享的 `http.Client{Timeout: 10s}`，`NewRequestWithContext(ctx)`，非 2xx 返回错误 | S |
| 4 | 中 | 手机号哈希是不加盐的 SHA-256（`Hash`/`GetDataHash`），手机号空间很小，可以穷举还原 | `crypto/aes.go` | 改成 HMAC-SHA256（独立密钥），提供迁移期双写方案 | M |
| 5 | 中 | `IP2Area` 走明文 HTTP 的 ip-api.com（免费版不能商用，且有限速），出错时静默返回空结构体 | `ip.go` | 改用 HTTPS 或离线库（ip2region），返回 error | S |
| 6 | 中 | 同一功能有三套排队实现：`queue.go`（r2day/db 全局）、`q.go`（open4go/db）、`r2id/ticket.go`、`p7/oqueue` | 跨仓库 | 保留 `p7/oqueue` 或 `q.go` 中的一套，其余标记 Deprecated | M |
| 7 | 中 | 同时依赖 `r2day/db` 和 `open4go/db` 两套连接池 | `queue.go`、`db/redis.go` | 统一用 `open4go/db` | S |
| 8 | 低 | `rand.Seed` 已废弃，每次调用都重新播种 | `avatar.go` | 删掉 Seed（Go 1.20+ 会自动播种），或改用 `math/rand/v2` | S |
| 9 | 低 | `Fen2Yuan` 先转 float64 再格式化，大金额或负数时可能有精度和取整问题 | `amount.go` | 用整数运算：`fmt.Sprintf("%d.%02d", a/100, abs(a%100))` | S |
| 10 | 低 | `GetDataHash` 和 `Hash`、`GetDataSuffix` 和 `Suffix` 功能重复 | `crypto/aes.go` | 保留一套 | S |
| 11 | 中 | govulncheck 发现 7 个问题：x/net v0.43.0、otel v1.38.0（GO-2026-5506）、docker（来自旧版 log） | `go.mod` | 升级 log、x/net、otel | S |
| 12 | 中 | 除了 crypto，其他代码都没有测试（库存 Lua 脚本、签名最需要测试） | — | 用 miniredis 测 `SellProduct` 的并发扣减；用第三方文档里的示例测签名 | M |
