package util9s

import (
	"fmt"
	"math/rand"
)

func RandNumber() string {
	// 生成0到100之间的随机数
	randomNumber := rand.Intn(101) // Intn(n)返回一个取值范围[0, n)的伪随机数
	return fmt.Sprintf("%d", randomNumber)

}

// GetRandomAvatar 获取随机头像
func GetRandomAvatar() string {
	return "https://avatar.iran.liara.run/public/" + RandNumber()
}
