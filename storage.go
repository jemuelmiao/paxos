package paxos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandlePutStorage(w http.ResponseWriter, r *http.Request) {
	value := r.URL.Query().Get("value")

	var proposalId string
	var proposalValue string
	var err error
	tk := time.NewTicker(50 * time.Millisecond)
	to := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-tk.C:
			proposalId, proposalValue, err = DefaultProposer.Prepare(value)
			if err != nil {
				break
			}
			err = DefaultProposer.Propose(proposalId, proposalValue)
			if err != nil {
				break
			}
			if proposalValue == value {
				WriteRsp(w, 0, "put storage success", proposalValue)
			} else {
				WriteRsp(w, -1, "put storage fail", proposalValue)
			}
			return
		case <-to.C:
			WriteRsp(w, -1, "put storage time out", nil)
			return
		}
	}
}

func HandleGetStorage(w http.ResponseWriter, r *http.Request) {
	header := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
	type GetValueRsp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}
	var values []string
	for _, host := range DefaultHosts {
		go func(h string) {
			url := fmt.Sprintf("http://%v/paxos/value", h)
			status, body, err := HttpGet(url, header, nil)
			if err != nil || status >= 400 {
				fmt.Printf("request propose fail, host:%v, err:%v, status:%v\n", h, err, status)
				return
			}
			var rsp GetValueRsp
			if err := json.Unmarshal(body, &rsp); err != nil {
				fmt.Printf("unmarshal fail, body:%v, err:%v\n", string(body), err)
				return
			}
			values = append(values, rsp.Data)
		}(host)
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	timeout := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			if len(values) > len(DefaultHosts)/2 {
				//收到超半数回复
				//判断是否有值
				countMap := make(map[string]int)
				for _, value := range values {
					if _, ok := countMap[value]; !ok {
						countMap[value] = 0
					}
					countMap[value] += 1
				}
				for value, count := range countMap {
					if count > len(DefaultHosts)/2 {
						WriteRsp(w, 0, "", value)
						return
					}
				}
			}
			if len(values) == len(DefaultHosts) {
				WriteRsp(w, -1, "has no value", nil)
				return
			}
		case <-timeout.C:
			//未收到超半数回复
			WriteRsp(w, -1, "get storage time out", nil)
			return
		}
	}
}
