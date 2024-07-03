package util9s

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

// Signature 签名算法
func Signature(key string, sendParamEntity interface{}) string {
	str := GetFieldString(sendParamEntity)
	if str == "" {
		return ""
	}
	stringA := fmt.Sprintf("%s&%s=%s", str, "key", key)
	stringA = strings.ToUpper(stringA)
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(stringA))
	sign := hex.EncodeToString(md5Ctx.Sum(nil))
	return str + "&sign=" + strings.ToUpper(sign)
}
