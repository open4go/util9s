package util9s

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open4go/log"
	"io"
	"net/http"
)

const (
	jsonContentType = "application/json"
)

// Post 请求
func Post(urlStr string, reqParam interface{}) ([]byte, error) {
	reqByte, err := json.Marshal(reqParam)
	if err != nil {
		logError(urlStr, err, "请求参数序列化失败")
		return nil, err
	}

	resp, err := http.Post(urlStr, jsonContentType, bytes.NewBuffer(reqByte))
	if err != nil {
		logError(urlStr, err, "Post 请求失败")
		return nil, err
	}
	defer closeResponseBody(resp.Body, urlStr)

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		logError(urlStr, err, "读取响应出错")
		return nil, err
	}
	return respByte, nil
}

// Get http请求
func Get(urlStr string, rsp interface{}) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		logError(urlStr, err, "Get 请求失败")
		return err
	}
	defer closeResponseBody(resp.Body, urlStr)

	if resp.StatusCode != 200 {
		err := errors.New("resp status code is no 200")
		logError(urlStr, err, "响应码错误")
		return err
	}

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		logError(urlStr, err, "读取响应出错")
		return err
	}

	if err := json.Unmarshal(respByte, rsp); err != nil {
		logError(urlStr, err, fmt.Sprintf("解析响应出错: %s", string(respByte)))
		return err
	}
	return nil
}

// FetchByPost 通过post 获取结果
func FetchByPost(urlStr string, reqParam interface{}, rsp interface{}) error {
	respBytes, err := Post(urlStr, reqParam)
	if err != nil {
		logError(urlStr, err, "Post 请求失败")
		return err
	}

	// 解析参数
	if err := json.Unmarshal(respBytes, rsp); err != nil {
		logError(urlStr, err, "解析响应出错")
		return err
	}
	return nil
}

func logError(url string, err error, message string) {
	log.Log().WithField("url", url).
		WithField("err", err.Error()).
		Error(message)
}

func closeResponseBody(body io.ReadCloser, url string) {
	if err := body.Close(); err != nil {
		log.Log().WithField("url", url).
			WithField("err", err.Error()).
			Error("关闭响应体出错")
	}
}
