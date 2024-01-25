package util9s

import (
	"bytes"
	"encoding/json"
	"github.com/open4go/log"
	"io"
	"net/http"
)

const (
	jsonContentType = "application/json"
)

// Post 请求
func Post(urlStr string, reqParam interface{}) ([]byte, error) {
	reqByte, _ := json.Marshal(reqParam)
	resp, err := http.Post(urlStr, jsonContentType, bytes.NewBuffer(reqByte))
	if err != nil {
		log.Log().WithField("url", urlStr).
			WithField("err", err.Error()).
			Error("Post 请求失败")
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Log().WithField("url", urlStr).
			WithField("err", err.Error()).Error("读取响应出错")
		return nil, err
	}
	return respByte, nil
}

// Get http请求
func Get(urlStr string, rsp interface{}) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		log.Log().WithField("url", urlStr).
			WithField("err", err.Error()).Error("Get 请求失败")
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Log().WithField("url", urlStr).
			WithField("err", err.Error()).Error("读取响应出错")
		return err
	}

	if err := json.Unmarshal(respByte, rsp); err != nil {
		log.Log().WithField("url", urlStr).WithField("err", err.Error()).
			WithField("respByte", respByte).Error("解析响应出错")
		return err
	}
	return nil
}
