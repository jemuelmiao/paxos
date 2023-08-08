package paxos

import "errors"

type Acceptor struct {
	state              string //状态，updating、active
	MinProposalId      string //能够同意的最小提案号
	AcceptedProposalId string //已接收的提案号
	AcceptedValue      string //已接收的提案值
}

func NewAcceptor() *Acceptor {
	return &Acceptor{
		state: "active",
	}
}

// acceptedProposalId, acceptedValue, error
func (acceptor *Acceptor) Promise(proposalId string) (string, string, error) {
	if proposalId > acceptor.MinProposalId {
		//这里根据状态判断是返回acceptor.AcceptedProposalId还是""
		acceptor.MinProposalId = proposalId
		if acceptor.state == "updating" {
			return acceptor.AcceptedProposalId, acceptor.AcceptedValue, nil
		} else {
			//未接收提案
			acceptor.state = "updating"
			return "", "", nil
		}
	} else {
		return "", "", errors.New("reject")
	}
}

func (acceptor *Acceptor) Accept(proposalId, proposalValue string) error {
	if proposalId >= acceptor.MinProposalId {
		acceptor.AcceptedProposalId = proposalId
		acceptor.AcceptedValue = proposalValue
		return nil
	} else {
		return errors.New("reject")
	}
}

func (acceptor *Acceptor) GetValue() string {
	return acceptor.AcceptedValue
}

func (acceptor *Acceptor) ResetState() {
	acceptor.state = "active"
}
