package paxos

import (
	"encoding/json"
	"io"
	"net/http"
)

var (
	DefaultProposer *Proposer
	DefaultAcceptor *Acceptor
	//DefaultLearner  *Learner
	DefaultHosts []string
)

func HandlePromise(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buff, err := io.ReadAll(r.Body)
	if err != nil {
		WriteRsp(w, -1, err.Error(), nil)
		return
	}
	type Req struct {
		ProposalId string `json:"proposal_id"`
	}
	var req Req
	if err := json.Unmarshal(buff, &req); err != nil {
		WriteRsp(w, -1, err.Error(), nil)
		return
	}
	acceptedProposalId, acceptedValue, err := DefaultAcceptor.Promise(req.ProposalId)
	if err != nil {
		WriteRsp(w, -1, err.Error(), nil)
		return
	}
	type Rsp struct {
		AcceptedProposalId string `json:"accepted_proposal_id"` //已接收的提案号
		AcceptedValue      string `json:"accepted_value"`       //已接收的提案值
	}
	rsp := Rsp{
		AcceptedProposalId: acceptedProposalId,
		AcceptedValue:      acceptedValue,
	}
	WriteRsp(w, 0, "", rsp)
}

func HandleAccept(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buff, err := io.ReadAll(r.Body)
	if err != nil {
		WriteRsp(w, -1, err.Error(), nil)
		return
	}
	type Req struct {
		ProposalId    string `json:"proposal_id"`
		ProposalValue string `json:"proposal_value"`
	}
	var req Req
	if err := json.Unmarshal(buff, &req); err != nil {
		WriteRsp(w, -1, err.Error(), nil)
		return
	}
	err = DefaultAcceptor.Accept(req.ProposalId, req.ProposalValue)
	if err != nil {
		WriteRsp(w, -1, err.Error(), nil)
		return
	}
	WriteRsp(w, 0, "", nil)
}

func HandleGetValue(w http.ResponseWriter, r *http.Request) {
	WriteRsp(w, 0, "", DefaultAcceptor.GetValue())
}

func HandleResetState(w http.ResponseWriter, r *http.Request) {
	DefaultAcceptor.ResetState()
	WriteRsp(w, 0, "", nil)
}

func init() {
	DefaultProposer = NewProposer("node-1")
	DefaultAcceptor = NewAcceptor()
	DefaultHosts = append(DefaultHosts, "172.10.0.1:18090")
	DefaultHosts = append(DefaultHosts, "172.10.0.2:18090")
	DefaultHosts = append(DefaultHosts, "172.10.0.3:18090")
}
