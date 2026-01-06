package util9s

import (
	"context"
	"fmt"
	r2redis "github.com/open4go/db/redis"
	"github.com/open4go/log"
	"github.com/redis/go-redis/v9"
	"time"
)

// GetRedisCacheHandler 获取数据库handler 这里定义一个方法
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
		ctx,
		GetRedisCacheHandler(ctx),
	}
}

// GenerateQueueNumber 下单时调用
func (q *MyQueue) GenerateQueueNumber(storeID string) (int64, error) {
	today := time.Now().Format("2006-01-02")
	queueKey := fmt.Sprintf("queue:counter:%s:%s", storeID, today)

	queueNumber, err := q.Rdb.Incr(q.Ctx, queueKey).Result()
	if err != nil {
		return 0, err
	}
	q.Rdb.Expire(q.Ctx, queueKey, time.Hour*48)
	return queueNumber, nil
}

// AddOrderToQueue 支付成功时写入
func (q *MyQueue) AddOrderToQueue(storeID string, orderNumber string) error {
	today := time.Now().Format("2006-01-02")
	queueKey := fmt.Sprintf("queue:%s:%s", storeID, today)

	_, err := q.Rdb.ZAdd(q.Ctx, queueKey, redis.Z{
		Score:  float64(time.Now().UnixNano()), // 使用当前时间戳作为排序依据，保证顺序
		Member: orderNumber,
	}).Result()
	if err != nil {
		return err
	}
	return nil
}

// CompleteOrder 完成交易时（例如：二维码扫描/或者移动pad订单
func (q *MyQueue) CompleteOrder(storeID string, orderNumber string) error {
	today := time.Now().Format("2006-01-02")
	queueKey := fmt.Sprintf("queue:%s:%s", storeID, today)

	_, err := q.Rdb.ZRem(q.Ctx, queueKey, orderNumber).Result()
	if err != nil {
		return err
	}

	return nil
}

// GetQueuePosition 查看订单详情时
func (q *MyQueue) GetQueuePosition(storeID string, orderNumber string) (int64, error) {
	today := time.Now().Format("2006-01-02")
	queueKey := fmt.Sprintf("queue:%s:%s", storeID, today)

	position, err := q.Rdb.ZRank(q.Ctx, queueKey, fmt.Sprintf("%s", orderNumber)).Result()
	if err != nil {
		return 0, err
	}

	return position + 1, nil
}

func (q *MyQueue) BatchGetQueuePosition(
	storeID string,
	orderNos []string,
) (map[string]int64, error) {

	today := time.Now().Format("2006-01-02")
	queueKey := fmt.Sprintf("queue:%s:%s", storeID, today)

	pipe := q.Rdb.Pipeline()
	cmds := make(map[string]*redis.IntCmd)

	for _, orderNo := range orderNos {
		cmds[orderNo] = pipe.ZRank(q.Ctx, queueKey, orderNo)
	}

	_, err := pipe.Exec(q.Ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(orderNos))
	for orderNo, cmd := range cmds {
		if v, err := cmd.Result(); err == nil {
			result[orderNo] = v + 1
		}
	}

	return result, nil
}
