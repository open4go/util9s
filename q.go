package util9s

import (
	"context"
	"errors"
	"fmt"
	"time"

	r2redis "github.com/open4go/db/redis"
	"github.com/open4go/log"
	"github.com/redis/go-redis/v9"
)

/*

功能,说明
生成排队号,下单时为门店生成当天递增的排队号
加入排队,支付成功后将订单加入排队列表
完成订单,出餐/核销后从排队列表移除
查询位置,查询单个或批量订单的当前排队位置
*/

const (
	// QueueTTL 队列相关 key 默认过期时间
	QueueTTL = 48 * time.Hour
)

// GetRedisCacheHandler 获取 Redis handler
func GetRedisCacheHandler(ctx context.Context) *redis.Client {
	handler, err := r2redis.DBPool.GetHandler("cache")
	if err != nil {
		log.Log(ctx).Fatal(err)
	}
	return handler
}

type MyQueue struct {
	Ctx context.Context
	Rdb *redis.Client
}

// MakeMyQueue 创建队列服务
func MakeMyQueue(ctx context.Context) MyQueue {
	return MyQueue{
		Ctx: ctx,
		Rdb: GetRedisCacheHandler(ctx),
	}
}

// queueCounterKey 当天排队号计数器
func queueCounterKey(storeID, date string) string {
	return fmt.Sprintf("queue:counter:%s:%s", storeID, date)
}

// queueKey 当天订单排队列表（ZSet）
func queueKey(storeID, date string) string {
	return fmt.Sprintf("queue:%s:%s", storeID, date)
}

// today 获取当天日期字符串
func today() string {
	return time.Now().Format("2006-01-02")
}

// GenerateQueueNumber 下单时调用，生成排队号
func (q *MyQueue) GenerateQueueNumber(storeID string) (int64, error) {
	key := queueCounterKey(storeID, today())

	pipe := q.Rdb.Pipeline()
	incrCmd := pipe.Incr(q.Ctx, key)
	pipe.Expire(q.Ctx, key, QueueTTL) // 确保 48 小时后自动删除

	if _, err := pipe.Exec(q.Ctx); err != nil {
		return 0, err
	}

	return incrCmd.Val(), nil
}

// AddOrderToQueue 支付成功时写入排队列表
func (q *MyQueue) AddOrderToQueue(storeID string, orderNumber string) error {
	key := queueKey(storeID, today())

	pipe := q.Rdb.Pipeline()
	pipe.ZAdd(q.Ctx, key, redis.Z{
		Score:  float64(time.Now().UnixNano()), // 纳秒时间戳保证顺序
		Member: orderNumber,
	})
	pipe.Expire(q.Ctx, key, QueueTTL) // 每次写入都续期，保证 48 小时后自动删除

	_, err := pipe.Exec(q.Ctx)
	return err
}

// CompleteOrder 完成订单时从队列移除
func (q *MyQueue) CompleteOrder(storeID string, orderNumber string) error {
	key := queueKey(storeID, today())

	// 移除后也续期一次（防止 key 很快被删掉影响其他查询，可选）
	pipe := q.Rdb.Pipeline()
	pipe.ZRem(q.Ctx, key, orderNumber)
	pipe.Expire(q.Ctx, key, QueueTTL)

	_, err := pipe.Exec(q.Ctx)
	return err
}

// GetQueuePosition 查询单个订单的排队位置（从 1 开始）
func (q *MyQueue) GetQueuePosition(storeID string, orderNumber string) (int64, error) {
	key := queueKey(storeID, today())

	pos, err := q.Rdb.ZRank(q.Ctx, key, orderNumber).Result()
	if errors.Is(err, redis.Nil) {
		// 不在队列中
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return pos + 1, nil
}

// BatchGetQueuePosition 批量查询订单排队位置
func (q *MyQueue) BatchGetQueuePosition(storeID string, orderNos []string) (map[string]int64, error) {
	if len(orderNos) == 0 {
		return map[string]int64{}, nil
	}

	key := queueKey(storeID, today())
	pipe := q.Rdb.Pipeline()
	cmds := make(map[string]*redis.IntCmd, len(orderNos))

	for _, orderNo := range orderNos {
		cmds[orderNo] = pipe.ZRank(q.Ctx, key, orderNo)
	}

	if _, err := pipe.Exec(q.Ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	result := make(map[string]int64, len(orderNos))
	for orderNo, cmd := range cmds {
		pos, err := cmd.Result()
		if errors.Is(err, redis.Nil) {
			result[orderNo] = 0 // 不在队列
			continue
		}
		if err != nil {
			return nil, err
		}
		result[orderNo] = pos + 1
	}

	return result, nil
}
