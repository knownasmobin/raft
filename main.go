package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

//Define the number of nodes
var raftCount = 3

//Node pool
var nodeTable map[string]string

//Election timeout time (unit: second)
var timeout = 3

//Heartbeat detection timeout time
var heartBeatTimeout = 7

//Heartbeat detection frequency (unit: second)
var heartBeatTimes = 3

//Used to store messages
var MessageStore = make(map[int]string)

func main() {
	//Define three nodes node numbers -monitoring port number
	nodeTable = map[string]string{
		"A": ":9000",
		"B": ":9001",
		"C": ":9002",
	}
	//Specify the node number when running the program
	if len(os.Args) < 1 {
		log.Fatal("程序参数不正确")
	}

	id := os.Args[1]
	//Pass in node number, port number, create RAFT instance
	raft := NewRaft(id, nodeTable[id])
	//Enable RPC, register RAFT
	go rpcRegister(raft)
	//Open heartbeat detection
	go raft.heartbeat()
	//Open a http monitoring
	if id == "A" {
		go raft.httpListen()
	}

Circle:
	//Start an election
	go func() {
		for {
			//Become a candidate node
			if raft.becomeCandidate() {
				//After becoming a post-elect node, ask for votes from other nodes to conduct elections
				if raft.election() {
					break
				} else {
					continue
				}
			} else {
				break
			}
		}
	}()

	//Heartbeat is measured
	for {
		//0.5 Detect once in a second
		time.Sleep(time.Millisecond * 5000)
		if raft.lastHeartBeartTime != 0 && (millisecond()-raft.lastHeartBeartTime) > int64(raft.timeout*1000) {
			fmt.Printf("The heartbeat detection timed out and has exceeded %d seconds\n", raft.timeout)
			fmt.Println("Elections are about to reopen")
			raft.reDefault()
			raft.setCurrentLeader("-1")
			raft.lastHeartBeartTime = 0
			goto Circle
		}
	}
}
