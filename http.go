package util9s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/open4go/log"
	"github.com/spf13/viper"
)

// InternalTokenHeader 服务间调用令牌请求头，与 open4go/middle.RPCAuthMiddleware 一致。
const InternalTokenHeader = "X-Internal-Token"

// internalToken 读取服务间令牌（rpc.auth_token，兼容旧配置 rpc.token）。
func internalToken() string {
	if t := strings.TrimSpace(viper.GetString("rpc.auth_token")); t != "" {
		return t
	}
	return strings.TrimSpace(viper.GetString("rpc.token"))
}

// isInternalRPC 仅对路径中含 /rpc/ 的地址（内部服务接口）携带令牌，避免把令牌发给第三方。
func isInternalRPC(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return strings.Contains(u.Path+"/", "/rpc/")
}

// doInternal 与 http.Get / http.Post 行为一致；访问内部 /rpc/ 接口时自动加上 X-Internal-Token。
func doInternal(method, urlStr, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, body) //nolint:noctx // 保持与原 http.Get/http.Post 相同的语义（不随请求 context 取消）
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if isInternalRPC(urlStr) {
		if t := internalToken(); t != "" {
			req.Header.Set(InternalTokenHeader, t)
		}
	}
	return http.DefaultClient.Do(req)
}

const (
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
)

// Post 请求
func Post(ctx context.Context, urlStr string, reqParam interface{}) ([]byte, error) {
	reqByte, err := json.Marshal(reqParam)
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).Error(err)
		return nil, err
	}

	resp, err := doInternal(http.MethodPost, urlStr, jsonContentType, bytes.NewBuffer(reqByte))
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).WithField("reqParam", reqParam).
			Error(err)
		return nil, err
	}
	defer closeResponseBody(ctx, resp.Body, urlStr)

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).
			WithField("reqParam", reqParam).
			WithField("respByte", string(respByte)).
			Error(err)
		return nil, err
	}
	return respByte, nil
}

// Get http请求
func Get(ctx context.Context, urlStr string, rsp interface{}) error {
	resp, err := doInternal(http.MethodGet, urlStr, "", nil)
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).Error(err)
		return err
	}
	defer closeResponseBody(ctx, resp.Body, urlStr)

	if resp.StatusCode != 200 {
		err := fmt.Errorf("resp status code is %d | %s",
			resp.StatusCode, resp.Status)
		log.Log(ctx).WithField("url", urlStr).WithField("resp", resp).
			Error(err)
		return err
	}

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).
			WithField("respByte", string(respByte)).
			Error(err)
		return err
	}

	if err := json.Unmarshal(respByte, rsp); err != nil {
		log.Log(ctx).WithField("url", urlStr).
			WithField("respByte", string(respByte)).
			Error(err)
		return err
	}
	return nil
}

// FetchByPost 通过post 获取结果
func FetchByPost(ctx context.Context, urlStr string, reqParam interface{}, rsp interface{}) error {
	respBytes, err := Post(ctx, urlStr, reqParam)
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).WithField("req", reqParam).
			Error(err)
		return err
	}

	// 解析参数
	if err := json.Unmarshal(respBytes, rsp); err != nil {
		log.Log(ctx).WithField("url", urlStr).WithField("req", reqParam).
			Error(err)
		return err
	}
	return nil
}

func closeResponseBody(ctx context.Context, body io.ReadCloser, url string) {
	if err := body.Close(); err != nil {
		log.Log(ctx).WithField("url", url).
			WithField("err", err.Error()).
			Error("关闭响应体出错")
	}
}

// PostForm 发送post请求
// 并且走签名
func PostForm(ctx context.Context, urlStr, secretKey string, reqParam any) ([]byte, error) {
	formData := Signature(secretKey, reqParam)
	formRequest := strings.NewReader(formData)
	resp, err := http.Post(urlStr, formContentType, formRequest)
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).WithField("reqParam", reqParam).
			Error(err)
		return nil, err
	}
	defer closeResponseBody(ctx, resp.Body, urlStr)
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).
			WithField("reqParam", reqParam).
			WithField("respByte", string(respByte)).
			Error(err)
		return nil, err
	}
	return respByte, nil
}
