package util9s

import (
	"context"
	"crypto/rsa"
	"github.com/open4go/log"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type WxConfig struct {
	PrivateKey *rsa.PrivateKey
}

var (
	// GlobalWxConfig 微信配置
	GlobalWxConfig WxConfig
)

// InitWxConfig 加载密钥
func InitWxConfig(ctx context.Context, path string) (WxConfig, error) {
	wxc := WxConfig{}
	// 加载商户私钥
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(path)
	if err != nil {
		log.Log(ctx).WithField("path", path).Error("加载商户密钥失败")
		return wxc, err
	}
	wxc.PrivateKey = mchPrivateKey
	// 全局
	GlobalWxConfig = wxc
	return wxc, nil
}

// LoadPrivateKey 加载密钥
func LoadPrivateKey(ctx context.Context, key string) (*rsa.PrivateKey, error) {
	// 加载商户私钥
	mchPrivateKey, err := utils.LoadPrivateKey(key)
	if err != nil {
		log.Log(ctx).WithField("key", key).Error("加载商户密钥失败")
		return nil, err
	}
	return mchPrivateKey, nil
}
