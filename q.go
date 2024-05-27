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
func GetRedisCacheHandler() *redis.Client {
	handler, err := r2redis.DBPool.GetHandler("cache")
	if err != nil {
		log.Log().Fatal(err)
	}
	return handler
}

type MyQueue struct {
	Ctx context.Context
	Rdb *redis.Client
}

// MakeMyQueue 创建队列服务
func MakeMyQueue() MyQueue {
	return MyQueue{
		context.Background(),
		GetRedisCacheHandler(),
	}
}

// GenerateQueueNumber 下单时调用
func (q *MyQueue) GenerateQueueNumber(storeID string) (int64, error) {
	today := time.Now().Format("2006-01-02")
	queueKey := fmt.Sprintf("queue:%s:%s", storeID, today)

	queueNumber, err := q.Rdb.Incr(q.Ctx, queueKey).Result()
	if err != nil {
		return 0, err
	}
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
func (q *MyQueue) GetQueuePosition(storeID string, orderNumber int64) (int64, error) {
	today := time.Now().Format("2006-01-02")
	queueKey := fmt.Sprintf("queue:%s:%s", storeID, today)

	position, err := q.Rdb.ZRank(q.Ctx, queueKey, fmt.Sprintf("%d", orderNumber)).Result()
	if err != nil {
		return 0, err
	}

	return position + 1, nil
}

//func main() {
//	rdb := newRedisClient()
//	storeID := "store123"
//
//	// 生成新的排队号
//	orderNumber, err := generateQueueNumber(rdb, storeID)
//	if err != nil {
//		fmt.Println("Error generating queue number:", err)
//		return
//	}
//
//	// 添加订单到队列
//	err = addOrderToQueue(rdb, storeID, orderNumber)
//	if err != nil {
//		fmt.Println("Error adding order to queue:", err)
//		return
//	}
//
//	// 获取订单排队位置
//	position, err := getQueuePosition(rdb, storeID, orderNumber)
//	if err != nil {
//		fmt.Println("Error getting queue position:", err)
//		return
//	}
//
//	fmt.Printf("Order %d is at position %d in the queue\n", orderNumber, position)
//
//	// 完成订单
//	err = completeOrder(rdb, storeID, orderNumber)
//	if err != nil {
//		fmt.Println("Error completing order:", err)
//		return
//	}
//
//	fmt.Printf("Order %d completed\n", orderNumber)
//}
