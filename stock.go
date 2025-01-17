package util9s

import (
	"context"
	"errors"
	"fmt"
	"github.com/open4go/util9s/db"
	"github.com/redis/go-redis/v9"
	"time"
)

// InitializeDailyStock 初始化每日库存
func InitializeDailyStock(ctx context.Context, productID string, maxSupply int, forceFresh bool) error {
	dateKey := fmt.Sprintf("product_stock:%s", time.Now().Format("2006-01-02"))

	if forceFresh {
		err := db.GetRedisCacheHandler(ctx).HDel(ctx, dateKey, productID).Err()
		if err != nil {
			return err
		}
	}

	// 检查键是否已经存在
	exists, err := db.GetRedisCacheHandler(ctx).HExists(ctx, dateKey, productID).Result()
	if err != nil {
		return err
	}

	// 如果键不存在，则初始化库存
	if !exists {
		err := db.GetRedisCacheHandler(ctx).HSet(ctx, dateKey, productID, maxSupply).Err()
		if err != nil {
			return err
		}
		// 设置键的过期时间为当天结束时间
		now := time.Now()
		expireTime := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		err = db.GetRedisCacheHandler(ctx).ExpireAt(ctx, dateKey, expireTime).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// SellProduct 销售商品
func SellProduct(ctx context.Context, productID string, quantity int) (string, error) {
	if quantity <= 0 {
		return "", errors.New("销售数量必须大于 0")
	}

	dateKey := fmt.Sprintf("product_stock:%s", time.Now().Format("2006-01-02"))

	// 检查商品是否存在库存
	exists, err := db.GetRedisCacheHandler(ctx).HExists(ctx, dateKey, productID).Result()
	if err != nil {
		return "", err
	}

	if !exists {
		return fmt.Sprintf("商品 %s 尚未上架！", productID), nil
	}

	// 使用 Lua 脚本原子性减少库存
	luaScript := `
		local stock = redis.call("HGET", KEYS[1], ARGV[1])
		if tonumber(stock) >= tonumber(ARGV[2]) then
			redis.call("HINCRBY", KEYS[1], ARGV[1], -tonumber(ARGV[2]))
			return tonumber(stock) - tonumber(ARGV[2])
		else
			return -1
		end
	`
	result, err := db.GetRedisCacheHandler(ctx).Eval(ctx, luaScript, []string{dateKey}, productID, quantity).Result()
	if err != nil {
		return "", err
	}

	remainingStock := int(result.(int64))
	if remainingStock == -1 {
		return fmt.Sprintf("商品 %s 库存不足，无法售出 %d 件！", productID, quantity), nil
	}
	return fmt.Sprintf("商品 %s 售出 %d 件，剩余库存: %d", productID, quantity, remainingStock), nil
}

// GetRemainingStock 查询当前剩余库存
func GetRemainingStock(ctx context.Context, productID string) (int, error) {
	dateKey := fmt.Sprintf("product_stock:%s", time.Now().Format("2006-01-02"))

	// 获取库存
	stock, err := db.GetRedisCacheHandler(ctx).HGet(ctx, dateKey, productID).Int()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return stock, nil
}

func GetRemainingStockNumber(ctx context.Context, productID string) int {
	stock, err := GetRemainingStock(ctx, productID)
	if err != nil {
		return 0
	} else {
		return stock
	}
}

// 测试示例
func demo() {
	ctx := context.TODO()
	// 初始化商品库存
	err := InitializeDailyStock(ctx, "product_1", 100, false)
	if err != nil {
		fmt.Println("初始化库存失败:", err)
		return
	}

	// 查询库存
	stockMsg, err := GetRemainingStock(ctx, "product_1")
	if err != nil {
		fmt.Println("查询库存失败:", err)
		return
	}
	fmt.Println(stockMsg)

	// 销售商品
	for i := 0; i < 105; i++ {
		sellMsg, err := SellProduct(ctx, "product_1", 1)
		if err != nil {
			fmt.Println("销售失败:", err)
			return
		}
		fmt.Println(sellMsg)
	}

	// 再次查询库存
	stockMsg, err = GetRemainingStock(ctx, "product_1")
	if err != nil {
		fmt.Println("查询库存失败:", err)
		return
	}
	fmt.Println(stockMsg)
}
