package paxos

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Rsp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:          2000,             // MaxIdleConns 表示空闲KeepAlive长连接数量，为0表示不限制
		MaxIdleConnsPerHost:   100,              // MaxIdleConnsPerHost 表示单Host空闲KeepAlive数量，系统默认为2
		ResponseHeaderTimeout: 60 * time.Second, // 等待后端响应头部的超时时间
		IdleConnTimeout:       90 * time.Second, // 长连接空闲超时回收时间
		TLSHandshakeTimeout:   10 * time.Second, // TLS握手超时时间
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	},
	Timeout: 120 * time.Second, // 整个请求超时时间
}

func HttpDo(url, host, method string, header map[string]string, body []byte, retry int) (int, []byte, error) {
	if header == nil {
		header = make(map[string]string)
	}
	if _, ok := header["Content-Type"]; !ok {
		header["Content-Type"] = "application/json; charset=utf-8"
	}
	var err error
	for i := 0; i < retry; i++ {
		var req *http.Request
		if body != nil {
			reqBody := bytes.NewBuffer(body)
			req, err = http.NewRequest(method, url, reqBody)
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		if len(host) != 0 {
			req.Host = host
		}
		for k, v := range header {
			req.Header.Set(k, v)
		}

		var rsp *http.Response
		rsp, err = httpClient.Do(req)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		defer rsp.Body.Close()

		rspBody, err := io.ReadAll(rsp.Body)
		if err != nil {
			return 0, nil, err
		}
		return rsp.StatusCode, rspBody, nil
	}
	return 0, nil, err

}

func marshalData(data interface{}) []byte {
	var buf []byte
	var ok bool
	if data != nil {
		if buf, ok = data.([]byte); !ok {
			buf, _ = json.Marshal(data)
		}
	}
	return buf
}

func HttpGet(url string, header map[string]string, data interface{}) (int, []byte, error) {
	buf := marshalData(data)
	return HttpDo(url, "", "GET", header, buf, 1)
}

func HttpPost(url string, header map[string]string, data interface{}) (int, []byte, error) {
	buf := marshalData(data)
	return HttpDo(url, "", "POST", header, buf, 1)
}

func HttpPut(url string, header map[string]string, data interface{}) (int, []byte, error) {
	buf := marshalData(data)
	return HttpDo(url, "", "PUT", header, buf, 1)
}

func HttpPatch(url string, header map[string]string, data interface{}) (int, []byte, error) {
	buf := marshalData(data)
	return HttpDo(url, "", "PATCH", header, buf, 1)
}

func HttpDelete(url string, header map[string]string, data interface{}) (int, []byte, error) {
	buf := marshalData(data)
	return HttpDo(url, "", "DELETE", header, buf, 1)
}

func WriteRsp(w http.ResponseWriter, code int, msg string, data interface{}) {
	rsp := &Rsp{
		Code: code,
		Msg:  msg,
		Data: data,
	}
	buf, _ := json.Marshal(rsp)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}
