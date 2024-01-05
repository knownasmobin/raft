package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/rpc"
)

// Waiting for node access
func (rf *Raft) getRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	//http://localhost:8080/req?message=ohmygod
	if len(request.Form["message"]) > 0 && rf.currentLeader != "-1" {
		message := request.Form["message"][0]
		m := new(Message)
		m.MsgID = getRandom()
		m.Msg = message
		//After receiving the message, forward it directly to the leader
		fmt.Println("HTTP supervised the message, ready to send to the leader, message ID:", m.MsgID)
		port := (nodeTable[rf.currentLeader])
		rp, err := rpc.DialHTTP("tcp", "127.0.0.1"+":"+port)
		if err != nil {
			log.Panic(err)
		}
		b := false
		err = rp.Call("Raft.LeaderReceiveMessage", m, &b)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("Whether the message has been sent to the leader：", b)
		writer.Write([]byte("ok!!!"))
	}
}

func (rf *Raft) httpListen() {
	//Create a GetRequest () callback method
	http.HandleFunc("/req", rf.getRequest)
	fmt.Println("Monitoring", httpPort, "port")
	if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
		fmt.Println(err)
		return
	}
}

// Return a ten -digit random number, as a message IDGIT
func getRandom() int {
	x := big.NewInt(10000000000)
	for {
		result, err := rand.Int(rand.Reader, x)
		if err != nil {
			log.Panic(err)
		}
		if result.Int64() > 1000000000 {
			return int(result.Int64())
		}
	}
}
