package main

import (
	"fmt"
	"net/http"
	"paxos"
)

func main() {
	http.HandleFunc("/paxos/promise", paxos.HandlePromise)
	http.HandleFunc("/paxos/accept", paxos.HandleAccept)
	http.HandleFunc("/paxos/value", paxos.HandleGetValue)
	http.HandleFunc("/paxos/reset", paxos.HandleResetState)
	//分布式锁
	http.HandleFunc("/lock/ensure", paxos.HandleEnsureLock)
	http.HandleFunc("/lock/release", paxos.HandleReleaseLock)
	//分布式存储
	http.HandleFunc("/storage/put", paxos.HandlePutStorage)
	http.HandleFunc("/storage/get", paxos.HandleGetStorage)
	//选举leader
	http.HandleFunc("/elect", paxos.HandleElect)
	http.HandleFunc("/leader", paxos.HandleLeader)
	if err := http.ListenAndServe(":18090", nil); err != nil {
		fmt.Println("start server fail:", err)
	}
}
