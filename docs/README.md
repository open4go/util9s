# util9s — 业务通用工具（金额、HTTP、签名、加密、库存、排队）

> 仓库：https://github.com/open4go/util9s ｜ 模块：`github.com/open4go/util9s` ｜ Go 1.24.0 ｜ 最近提交：2026-09-10（最新 tag v0.0.31）

## 项目简介
零散的业务工具函数：分转元、HTTP GET/POST（带日志）、MD5 表单签名、微信商户私钥加载、IP 归属地查询、经纬度距离、随机头像、Redis 日库存（Lua 原子扣减）、门店排队号，以及 `crypto` 子包（AES-GCM 加解密、手机号脱敏和哈希）。

## 目录结构
```
util9s/
├── amount.go   Fen2Yuan / FenS2Yuan / FenSF2Yuan
├── http.go     Post / Get / FetchByPost / PostForm（签名表单）
├── sign.go covert.go  Signature / GetFieldString（反射拼接 k=v&...，MD5 大写）
├── private.go  InitWxConfig / LoadPrivateKey（wechatpay-go）
├── ip.go       IP2Area（http://ip-api.com）
├── geo.go      CalcDistanceMeters
├── avatar.go   GetRandomAvatar
├── stock.go    InitializeDailyStock / SellProduct(Lua) / ReplenishStock / GetRemainingStock / GetHandoverStockStatus
├── q.go        MyQueue：GenerateQueueNumber / AddOrderToQueue / CompleteOrder / GetQueuePosition（open4go/db "cache"）
├── queue.go    StoreQueue：Push / MyIndex / Pop（r2day/db 全局 RDB）
├── db/redis.go GetRedisCacheHandler
└── crypto/     aes.go（SetPrimary / InitFromConfig / Encrypt / Decrypt / Hash / Suffix …）+ aes_test.go
```

## 使用示例
```go
// req5rsp/rpc/updateLevel.go
result, err := util9s.Post(ctx, url, payload)
// 库存
_ = util9s.InitializeDailyStock(ctx, mid, sid, pid, 100, false)
msg, err := util9s.SellProduct(ctx, mid, sid, pid, 2)
// 加密（启动时）
if err := crypto.InitFromConfig(); err != nil { ... } // PHONE_AES_KEY 或 encryptKey.data
enc, _ := crypto.Encrypt(phone)
```

## 依赖
`go-redis/v9 v9.7.3`、`wechatpay-go`、`viper`、`open4go/{db v0.0.13, log v0.0.18}`、`r2day/db v0.3.5`

## 构建与测试
`go build`、`go vet` 通过；`go test ./...` 通过（只有 `crypto/aes_test.go`）。

## 与其他仓库的关系
被 `req5rsp`（rpc 包）依赖；排队功能和 `r2id/ticket.go`、`p7/oqueue` 重复；加密功能和 `rest/secret.go` 重复。
