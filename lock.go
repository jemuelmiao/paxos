package paxos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleEnsureLock(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

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
				//判断是否有锁
				countMap := make(map[string]int)
				for _, value := range values {
					if _, ok := countMap[value]; !ok {
						countMap[value] = 0
					}
					countMap[value] += 1
				}
				for value, count := range countMap {
					if count > len(DefaultHosts)/2 && value != "" {
						WriteRsp(w, -1, "has already lock", value)
						return
					}
				}
				//加锁
				var proposalId string
				var proposalValue string
				var err error
				tk := time.NewTicker(50 * time.Millisecond)
				to := time.NewTimer(5 * time.Second)
				for {
					select {
					case <-tk.C:
						proposalId, proposalValue, err = DefaultProposer.Prepare(key)
						if err != nil {
							break
						}
						err = DefaultProposer.Propose(proposalId, proposalValue)
						if err != nil {
							break
						}
						if proposalValue == key {
							WriteRsp(w, 0, "ensure lock success", proposalValue)
						} else {
							WriteRsp(w, -1, "ensure lock fail", proposalValue)
						}
						return
					case <-to.C:
						WriteRsp(w, -1, "ensure lock time out", nil)
						return
					}
				}
			}
		case <-timeout.C:
			//未收到超半数回复
			WriteRsp(w, -1, "ensure lock time out", nil)
			return
		}
	}
}

func HandleReleaseLock(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

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
				//判断是否有锁
				countMap := make(map[string]int)
				for _, value := range values {
					if _, ok := countMap[value]; !ok {
						countMap[value] = 0
					}
					countMap[value] += 1
				}
				for value, count := range countMap {
					if count > len(DefaultHosts)/2 {
						if value == key {
							//释放锁
							var proposalId string
							var proposalValue string
							var err error
							tk := time.NewTicker(50 * time.Millisecond)
							to := time.NewTimer(5 * time.Second)
							for {
								select {
								case <-tk.C:
									proposalId, proposalValue, err = DefaultProposer.Prepare("")
									if err != nil {
										break
									}
									err = DefaultProposer.Propose(proposalId, proposalValue)
									if err != nil {
										break
									}
									if proposalValue == "" {
										WriteRsp(w, 0, "release lock success", proposalValue)
									} else {
										WriteRsp(w, -1, "release lock fail", proposalValue)
									}
									//WriteRsp(w, 0, "release lock success", proposalValue)
									return
								case <-to.C:
									WriteRsp(w, -1, "release lock time out", nil)
									return
								}
							}
						} else {
							WriteRsp(w, -1, "can not release other lock", value)
							return
						}
					}
				}
			}
			if len(values) == len(DefaultHosts) {
				WriteRsp(w, -1, "has no lock", nil)
				return
			}
		case <-timeout.C:
			//未收到超半数回复
			WriteRsp(w, -1, "ensure lock time out", nil)
			return
		}
	}
}
