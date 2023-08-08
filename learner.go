package paxos

type Learner struct {
	AcceptedProposalId string //已接收的提案号
	AcceptedValue      string //已接收的提案值
}

func (learner *Learner) Learn(acceptedProposalId, acceptedValue string) {

}
