package util9s

import (
	"context"
	"errors"
	"fmt"
	"github.com/open4go/util9s/db"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	// AlertThreshold 库存告急阈值（可根据业务调整）
	AlertThreshold = 10
)

// getDateKey 生成当日库存 key
func getDateKey(merchantId, storeId string) string {
	return fmt.Sprintf("product_stock:%s:%s:%s", merchantId, storeId, time.Now().Format("2006-01-02"))
}

// getAlertSetKey 库存告急集合
func getAlertSetKey(merchantId, storeId string) string {
	return fmt.Sprintf("stock_alert:%s:%s:%s", merchantId, storeId, time.Now().Format("2006-01-02"))
}

// getZeroSetKey 零库存集合
func getZeroSetKey(merchantId, storeId string) string {
	return fmt.Sprintf("stock_zero:%s:%s:%s", merchantId, storeId, time.Now().Format("2006-01-02"))
}

// InitializeDailyStock 初始化每日库存
func InitializeDailyStock(ctx context.Context, merchantId string, storeId string, productID string, maxSupply int, forceFresh bool) error {
	dateKey := getDateKey(merchantId, storeId)

	if forceFresh {
		_ = db.GetRedisCacheHandler(ctx).HDel(ctx, dateKey, productID)
		// 补货时清理告急和零库存记录
		cleanStockStatus(ctx, merchantId, storeId, productID)
	}

	exists, err := db.GetRedisCacheHandler(ctx).HExists(ctx, dateKey, productID).Result()
	if err != nil {
		return err
	}

	if !exists {
		err = db.GetRedisCacheHandler(ctx).HSet(ctx, dateKey, productID, maxSupply).Err()
		if err != nil {
			return err
		}

		// 设置过期时间（7天后自动清理）
		now := time.Now()
		expireTime := time.Date(now.Year(), now.Month(), now.Day()+7, 0, 0, 0, 0, now.Location())
		db.GetRedisCacheHandler(ctx).ExpireAt(ctx, dateKey, expireTime)
	}

	return nil
}

// cleanStockStatus 补货或更新库存时清理告急/零库存状态
func cleanStockStatus(ctx context.Context, merchantId, storeId, productID string) {
	alertKey := getAlertSetKey(merchantId, storeId)
	zeroKey := getZeroSetKey(merchantId, storeId)

	pipe := db.GetRedisCacheHandler(ctx).Pipeline()
	pipe.SRem(ctx, alertKey, productID)
	pipe.SRem(ctx, zeroKey, productID)
	_, _ = pipe.Exec(ctx)
}

// ReplenishStock 补货（推荐使用此函数更新库存）
func ReplenishStock(ctx context.Context, merchantId, storeId, productID string, addQuantity int) error {
	if addQuantity <= 0 {
		return errors.New("补货数量必须大于0")
	}

	dateKey := getDateKey(merchantId, storeId)

	// 增加库存
	newStock, err := db.GetRedisCacheHandler(ctx).HIncrBy(ctx, dateKey, productID, int64(addQuantity)).Result()
	if err != nil {
		return err
	}

	// 补货后清理告急和零库存集合
	cleanStockStatus(ctx, merchantId, storeId, productID)

	fmt.Printf("商品 %s 补货 %d 件，当前库存: %d\n", productID, addQuantity, newStock)
	return nil
}

// SellProduct 销售商品
func SellProduct(ctx context.Context, merchantId string, storeId string, productID string, quantity int) (string, error) {
	if quantity <= 0 {
		return "", errors.New("销售数量必须大于 0")
	}

	dateKey := getDateKey(merchantId, storeId)

	exists, err := db.GetRedisCacheHandler(ctx).HExists(ctx, dateKey, productID).Result()
	if err != nil {
		return "", err
	}
	if !exists {
		return fmt.Sprintf("商品 %s 尚未上架！", productID), nil
	}

	// Lua 脚本原子扣减库存
	luaScript := `
        local stock = redis.call("HGET", KEYS[1], ARGV[1])
        if tonumber(stock) >= tonumber(ARGV[2]) then
            local newStock = redis.call("HINCRBY", KEYS[1], ARGV[1], -tonumber(ARGV[2]))
            return newStock
        else
            return -1
        end
    `

	result, err := db.GetRedisCacheHandler(ctx).Eval(ctx, luaScript, []string{dateKey}, productID, quantity).Result()
	if err != nil {
		return "", err
	}

	remaining := int(result.(int64))
	if remaining == -1 {
		return fmt.Sprintf("商品 %s 库存不足，无法售出 %d 件！", productID, quantity), nil
	}

	// === 新增：库存状态处理 ===
	alertKey := getAlertSetKey(merchantId, storeId)
	zeroKey := getZeroSetKey(merchantId, storeId)

	if remaining == 0 {
		db.GetRedisCacheHandler(ctx).SAdd(ctx, zeroKey, productID)
		db.GetRedisCacheHandler(ctx).SRem(ctx, alertKey, productID) // 从告急中移除
	} else if remaining <= AlertThreshold {
		db.GetRedisCacheHandler(ctx).SAdd(ctx, alertKey, productID)
	}

	return fmt.Sprintf("商品 %s 售出 %d 件，剩余库存: %d", productID, quantity, remaining), nil
}

// GetRemainingStock 查询剩余库存
func GetRemainingStock(ctx context.Context, merchantId string, storeId string, productID string) (int, error) {
	dateKey := getDateKey(merchantId, storeId)
	stock, err := db.GetRedisCacheHandler(ctx).HGet(ctx, dateKey, productID).Int()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return stock, nil
}

func GetRemainingStockNumber(ctx context.Context, merchantId string, storeId string, productID string) int {
	stock, _ := GetRemainingStock(ctx, merchantId, storeId, productID)
	return stock
}

// GetHandoverStockStatus 交接班库存状态查询
// 返回值：
//   - alertList : 库存告急商品列表（剩余 ≤ 10）
//   - zeroList  : 零库存商品列表
func GetHandoverStockStatus(ctx context.Context, merchantId, storeId string) ([]string, []string, error) {
	alertKey := getAlertSetKey(merchantId, storeId)
	zeroKey := getZeroSetKey(merchantId, storeId)

	// 获取告急商品
	alertList, err := db.GetRedisCacheHandler(ctx).SMembers(ctx, alertKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, nil, err
	}

	// 获取零库存商品
	zeroList, err := db.GetRedisCacheHandler(ctx).SMembers(ctx, zeroKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, nil, err
	}

	return alertList, zeroList, nil
}
