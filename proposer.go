package paxos

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type Proposer struct {
	ServerId string //服务id
}

func NewProposer(serverId string) *Proposer {
	return &Proposer{ServerId: serverId}
}

// 保证不重复
func (proposer *Proposer) GetProposalId() string {
	return fmt.Sprintf("%v_%v", time.Now().UnixNano(), proposer.ServerId)
}

// proposalId, proposalValue, error
func (proposer *Proposer) Prepare(proposalValue string) (string, string, error) {
	proposalId := proposer.GetProposalId()

	header := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
	data := map[string]interface{}{
		"proposal_id": proposalId,
	}
	type PrepareRsp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			AcceptedProposalId string `json:"accepted_proposal_id"` //已接收的提案号
			AcceptedValue      string `json:"accepted_value"`       //已接收的提案值
		} `json:"data"`
	}

	var receive [][2]string
	var reject int32
	for _, host := range DefaultHosts {
		go func(h string) {
			url := fmt.Sprintf("http://%v/paxos/promise", h)
			status, body, err := HttpGet(url, header, data)
			if err != nil || status >= 400 {
				fmt.Printf("request prepare fail, host:%v, err:%v, status:%v\n", h, err, status)
				return
			}
			var rsp PrepareRsp
			if err := json.Unmarshal(body, &rsp); err != nil {
				fmt.Printf("unmarshal fail, body:%v, err:%v\n", string(body), err)
				return
			}
			//acceptor已有更大提案号，拒绝
			if rsp.Code != 0 {
				fmt.Printf("prepare fail, code:%v, msg:%v\n", rsp.Code, rsp.Msg)
				atomic.AddInt32(&reject, 1)
				return
			}
			receive = append(receive, [2]string{rsp.Data.AcceptedProposalId, rsp.Data.AcceptedValue})
		}(host)
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	timeout := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			if len(receive) > len(DefaultHosts)/2 {
				//收到超半数回复promise
				var maxAcceptedProposalId string
				var maxAcceptedValue string
				for _, reply := range receive {
					if reply[0] != "" && reply[0] > maxAcceptedProposalId {
						maxAcceptedProposalId = reply[0]
						maxAcceptedValue = reply[1]
					}
				}
				//acceptor没有接收提案号，使用自己的值
				if maxAcceptedProposalId == "" {
					return proposalId, proposalValue, nil
				}
				//acceptor有提案值，取最大提案号的值
				return proposalId, maxAcceptedValue, nil
			} else if atomic.LoadInt32(&reject) > int32(len(DefaultHosts)/2) {
				//收到超半数拒绝
				return "", "", errors.New("receive more than half of acceptor reject")
			}
		case <-timeout.C:
			//未收到超半数回复promise
			return "", "", errors.New("not receive more than half of acceptor promise")
		}
	}
}

func (proposer *Proposer) Propose(proposalId, proposalValue string) error {
	header := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
	data := map[string]interface{}{
		"proposal_id":    proposalId,
		"proposal_value": proposalValue,
	}
	type ProposeRsp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	var replyCount int32
	var reject int32
	for _, host := range DefaultHosts {
		go func(h string) {
			if atomic.LoadInt32(&reject) > 0 {
				return
			}
			url := fmt.Sprintf("http://%v/paxos/accept", h)
			status, body, err := HttpGet(url, header, data)
			if err != nil || status >= 400 {
				fmt.Printf("request propose fail, host:%v, err:%v, status:%v\n", h, err, status)
				return
			}
			var rsp ProposeRsp
			if err := json.Unmarshal(body, &rsp); err != nil {
				fmt.Printf("unmarshal fail, body:%v, err:%v\n", string(body), err)
				return
			}
			//acceptor已有更大提案号，拒绝
			if rsp.Code != 0 {
				fmt.Printf("propose fail, code:%v, msg:%v\n", rsp.Code, rsp.Msg)
				atomic.AddInt32(&reject, 1)
				return
			}
			atomic.AddInt32(&replyCount, 1)
		}(host)
	}
	//有拒绝
	if atomic.LoadInt32(&reject) > 0 {
		return errors.New("some acceptor reject")
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	timeout := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&replyCount) > int32(len(DefaultHosts)/2) {
				//收到超半数回复，通过
				//重置所有acceptor状态为新一轮
				for _, host := range DefaultHosts {
					go func(h string) {
						url := fmt.Sprintf("http://%v/paxos/reset", h)
						HttpGet(url, header, nil)
					}(host)
				}
				return nil
			}
		case <-timeout.C:
			//未收到超半数回复
			return errors.New("not receive more than half of acceptor response")
		}
	}
}
