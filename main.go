package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type Data struct {
	Nodes            map[string]string `json:"nodes"`
	Timeout          int               `json:"Timeout"`
	HeartBeatTimeout int               `json:"heartBeatTimeout"`
	HeartBeatTimes   int               `json:"heartBeatTimes"`
	HttpPort         string            `json:"httpPort"`
}

// Define the number of nodes
var raftCount int

// Node pool
var nodeTable map[string]string

// Election timeout time (unit: second)
var timeout int

// Heartbeat detection timeout time
var heartBeatTimeout int

// Heartbeat detection frequency (unit: second)
var heartBeatTimes int

// Used to store messages
var MessageStore = make(map[int]string)

// HTTP port to listen
var httpPort string

func main() {
	// Read Nodes id and port number from json file

	data, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatal("Error when reading config.json: ", err)
	}
	var config Data
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	raftCount = len(config.Nodes)
	timeout = config.Timeout
	heartBeatTimeout = config.HeartBeatTimeout
	heartBeatTimes = config.HeartBeatTimes
	httpPort = config.HttpPort
	// Populate the nodeTable map
	nodeTable = config.Nodes
	
	fmt.Println("The number of nodes is:", raftCount)
	// Check if the number of nodes is odd
	if (raftCount % 2) != 1 {
		log.Fatalf("The number of nodes must be odd %d", raftCount) // Fatal is equivalent to Print() followed by a call to os.Exit(1).
	}
	// Specify the node number when running the program
	if len(os.Args) < 1 {
		log.Fatal("The program parameters are incorrect ")
	}

	id := os.Args[1]
	// Pass in node number, port number, create RAFT instance
	raft := NewRaft(id, nodeTable[id])
	// Enable RPC, register RAFT
	go rpcRegister(raft)
	// Open heartbeat detection
	go raft.heartbeat()
	// Open an HTTP monitoring
	if id == "A" {
		go raft.httpListen()
	}

	go startElection(raft)

	//Heartbeat is measured
	for {
		// 0.5 Detect once in a second
		time.Sleep(time.Millisecond * 5000)
		if raft.lastHeartBeartTime != 0 && (millisecond()-raft.lastHeartBeartTime) > int64(raft.timeout*1000) {
			fmt.Printf("The heartbeat detection timed out and has exceeded %d seconds\n", raft.timeout)
			fmt.Println("Elections are about to reopen")
			raft.reDefault()
			raft.setCurrentLeader("-1")
			raft.lastHeartBeartTime = 0
			go startElection(raft)
		}
	}
}

func startElection(raft *Raft) {
	fmt.Println("Start the election")

	// Define the election function
	election := func() {
		for {
			fmt.Println("Start the election timeout timer")
			// Become a candidate node
			if raft.becomeCandidate() {
				// After becoming a post-elect node, ask for votes from other nodes to conduct elections
				fmt.Println("become Candidate")
				if raft.election() {
					fmt.Println("Raft Election")
					break
				} else {
					fmt.Println("break election")
					continue
				}
			} else {
				fmt.Println("becom")
				break
			}
		}
	}

	// Start the election timeout timer
	for {
		// 0.5 Detect once in a second
		time.Sleep(time.Millisecond * 5000)
		if raft.lastHeartBeartTime != 0 && (millisecond()-raft.lastHeartBeartTime) > int64(raft.timeout*1000) {
			fmt.Printf("The heartbeat detection timed out and has exceeded %d seconds\n", raft.timeout)
			fmt.Println("Elections are about to reopen")
			raft.reDefault()
			raft.setCurrentLeader("-1")
			raft.lastHeartBeartTime = 0
			go election()
		}
	}
}
