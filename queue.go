package util9s

import (
	"context"
	"github.com/r2day/db"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	// StoreQueuePrefix 门店
	StoreQueuePrefix = "queue:store:"
)

// StoreQueue 门店排队
type StoreQueue struct {
	Ctx       context.Context
	StoreName string
}

// NewStoreQueue 创建新的门店缓存
func NewStoreQueue(ctx context.Context, storeName string) *StoreQueue {
	return &StoreQueue{
		Ctx:       ctx,
		StoreName: storeName,
	}
}

// Push 将门店信息存入缓存中
func (s *StoreQueue) Push(orderId string) error {
	score := time.Now().Unix()
	// 存入数据库 zadd store_queue:sc001 12347 order_003
	err := db.RDB.ZAdd(s.Ctx, StoreQueuePrefix+s.StoreName, redis.Z{
		Score:  float64(score),
		Member: orderId,
	}).Err()
	if err != nil {
		return err
	}
	return nil
}

// MyIndex 获取当前排队位置
// zrank store_queue:sc001 order_002
func (s *StoreQueue) MyIndex(orderId string) (int64, error) {
	// 存入数据库 zadd store_queue:sc001 12347 order_003
	queueNumber, err := db.RDB.ZRank(s.Ctx, StoreQueuePrefix+s.StoreName, orderId).Result()
	if err != nil {
		return 0, err
	}
	return queueNumber, nil
}

// Pop 当前订单完成后
// zrank store_queue:sc001 order_002
func (s *StoreQueue) Pop(orderId string) error {
	// 存入数据库 zrem store_queue:sc001 order_001
	err := db.RDB.ZRem(s.Ctx, StoreQueuePrefix+s.StoreName, orderId).Err()
	if err != nil {
		return err
	}
	return nil
}
