package util9s

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open4go/log"
	"io"
	"net/http"
	"strings"
)

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

	resp, err := http.Post(urlStr, jsonContentType, bytes.NewBuffer(reqByte))
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
	resp, err := http.Get(urlStr)
	if err != nil {
		log.Log(ctx).WithField("url", urlStr).Error(err)
		return err
	}
	defer closeResponseBody(ctx, resp.Body, urlStr)

	if resp.StatusCode != 200 {
		err := errors.New(fmt.Sprintf("resp status code is %d | %s",
			resp.StatusCode, resp.Status))
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
