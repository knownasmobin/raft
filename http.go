package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

// Waiting for node access
func (rf *Raft) getRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	if len(request.Form["message"]) > 0 && rf.currentLeader != "-1" {
		message := request.Form["message"][0]
		m := new(Message)
		m.MsgID = getRandom()
		m.Msg = message
		fmt.Println("HTTP supervised the message, ready to send to the leader, message ID:", m.MsgID)
		port := (nodeTable[rf.currentLeader])

		go func() {
			rp, err := rpc.DialHTTP("tcp", "127.0.0.1"+":"+port)
			if err != nil {
				log.Println(err)
				http.Error(writer, "Failed to connect to leader", http.StatusInternalServerError)
				return
			}
			defer rp.Close()

			b := false
			err = rp.Call("Raft.LeaderReceiveMessage", m, &b)
			if err != nil {
				log.Println(err)
				http.Error(writer, "Failed to send message to leader", http.StatusInternalServerError)
				return
			}
			fmt.Println("Whether the message has been sent to the leader：", b)
			writer.Write([]byte("ok!!!"))
		}()
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
