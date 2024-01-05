package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"time"
)

//RPC service registration
func rpcRegister(raft *Raft) {
	//Register a server
	err := rpc.Register(raft)
	if err != nil {
		log.Panic(err)
	}
	port := raft.node.Port
	//Bind the service to the HTTP protocol
	rpc.HandleHTTP()
	//Listening port
	err = http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Failure to register RPC service failed", err)
	}
}

func (rf *Raft) broadcast(method string, args interface{}, fun func(ok bool)) {
	//Set up and don't broadcast yourself for yourself
	for nodeID, port := range nodeTable {
		if nodeID == rf.node.ID {
			continue
		}
		//Connect remote RPC
		rp, err := rpc.DialHTTP("tcp", "127.0.0.1"+port)
		if err != nil {
			fun(false)
			continue
		}

		var bo = false
		err = rp.Call(method, args, &bo)
		if err != nil {
			fun(false)
			continue
		}
		fun(bo)
	}
}

//vote
func (rf *Raft) Vote(node NodeInfo, b *bool) error {
	if rf.votedFor != "-1" || rf.currentLeader != "-1" {
		*b = false
	} else {
		rf.setVoteFor(node.ID)
		fmt.Printf("The voting is successful, and the %s node has been voted\n", node.ID)
		*b = true
	}
	return nil
}

//Confirmation leader
func (rf *Raft) ConfirmationLeader(node NodeInfo, b *bool) error {
	rf.setCurrentLeader(node.ID)
	*b = true
	fmt.Println("Leading nodes in the network have been found，", node.ID, "Become leader！\n")
	rf.reDefault()
	return nil
}

//Heartbeat Test Replies
func (rf *Raft) HeartbeatRe(node NodeInfo, b *bool) error {
	rf.setCurrentLeader(node.ID)
	rf.lastHeartBeartTime = millisecond()
	fmt.Printf("Receive from leadership node %s Heartbeat detection\n", node.ID)
	fmt.Printf("当前时间为:%d\n\n", millisecond())
	*b = true
	return nil
}

//The follower node is used to receive the message, and then stores it in the message pool.
func (rf *Raft) ReceiveMessage(message Message, b *bool) error {
	fmt.Printf("Receive the information from the leader node, the ID is:%d\n", message.MsgID)
	MessageStore[message.MsgID] = message.Msg
	*b = true
	fmt.Println("I have replied to receive the message, and the leaders are printed after confirming")
	return nil
}

//The feedback of the follower node was confirmed by the leader node, and began to print the message
func (rf *Raft) ConfirmationMessage(message Message, b *bool) error {
	go func() {
		for {
			if _, ok := MessageStore[message.MsgID]; ok {
				fmt.Printf("raft Verification passes, you can print the message, the ID is：%d\n", message.MsgID)
				fmt.Println("News：", MessageStore[message.MsgID], "\n")
				rf.lastMessageTime = millisecond()
				break
			} else {
				//如果没有此消息，等一会看看!!!
				time.Sleep(time.Millisecond * 10)
			}

		}
	}()
	*b = true
	return nil
}

//The news of the leader received, the follower of the follower node forwarded
func (rf *Raft) LeaderReceiveMessage(message Message, b *bool) error {
	fmt.Printf("The leader node receives the news of the forwarding，id为:%d\n", message.MsgID)
	MessageStore[message.MsgID] = message.Msg
	*b = true
	fmt.Println("Ready to broadcast the message...")
	num := 0
	go rf.broadcast("Raft.ReceiveMessage", message, func(ok bool) {
		if ok {
			num++
		}
	})

	for {
		//自己默认收到了消息，所以减去一
		if num > raftCount/2-1 {
			fmt.Printf("More than half of the nodes have received message ID:%d \n RAFT verification passed, and the message can be printed. The ID is: %d\n", message.MsgID, message.MsgID)
			fmt.Println("News：", MessageStore[message.MsgID], "\n")
			rf.lastMessageTime = millisecond()
			fmt.Println("Ready to send messages to the client...")
			go rf.broadcast("Raft.ConfirmationMessage", message, func(ok bool) {
			})
			break
		} else {
			//Rest
			time.Sleep(time.Millisecond * 100)
		}
	}
	return nil
}
