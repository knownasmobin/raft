package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

//Declaration Raft node type
type Raft struct {
	node *NodeInfo
	//Number of voting obtained by this node
	vote int
	//Thread lock
	lock sync.Mutex
	//Node number
	me string
	//Current term
	currentTerm int
	//Vote for which node
	votedFor string
	//Current node status
	//0Follower1Candidate2Leader
	state int
	//The time to send the last message
	lastMessageTime int64
	//The time to send the last message
	lastHeartBeartTime int64
	//Leadership of the current node
	currentLeader string
	//HeartbeatTimeoutTime (unit:Second)
	timeout int
	//Receive voting success channel
	voteCh chan bool
	//Heartbeat signal
	heartBeat chan bool
}

type NodeInfo struct {
	ID   string
	Port string
}

type Message struct {
	Msg   string
	MsgID int
}

func NewRaft(id, port string) *Raft {
	node := new(NodeInfo)
	node.ID = id
	node.Port = port

	rf := new(Raft)
	//Node information
	rf.node = node
	//The current node obtains votes
	rf.setVote(0)
	//serial number
	rf.me = id
	//VoteForThreeNodesOf012,NoOneWillVoteForAnyone
	rf.setVoteFor("-1")
	//0Follower
	rf.setStatus(0)
	//The last heartbeat detection time
	rf.lastHeartBeartTime = 0
	rf.timeout = heartBeatTimeout
	//No leader at first
	rf.setCurrentLeader("-1")
	//Setting term
	rf.setTerm(0)
	//Voting channel
	rf.voteCh = make(chan bool)
	//Heartbeat
	rf.heartBeat = make(chan bool)
	return rf
}

//Modify nodes as candidate status
func (rf *Raft) becomeCandidate() bool {
	r := randRange(1500, 5000)
	//After the dormant random time, start becoming a candidate
	time.Sleep(time.Duration(r) * time.Millisecond)
	//If you find that this node has been voted or has a leader, you don’t have to transform the candidate status
	if rf.state == 0 && rf.currentLeader == "-1" && rf.votedFor == "-1" {
		//Turn node status to 1
		rf.setStatus(1)
		//Which node is set to vote
		rf.setVoteFor(rf.me)
		//Node term of term plus 1
		rf.setTerm(rf.currentTerm + 1)
		//There is no leader at present
		rf.setCurrentLeader("-1")
		//Vote for yourself
		rf.voteAdd()
		fmt.Println("This node has been changed to candidate status")
		fmt.Printf("Current number of votes：%d\n", rf.vote)
		//Enter the election channel
		return true
	} else {
		return false
	}
}

//Election
func (rf *Raft) election() bool {
	fmt.Println("Start the leaders election and broadcast to other nodes")
	go rf.broadcast("Raft.Vote", rf.node, func(ok bool) {
		rf.voteCh <- ok
	})
	for {
		select {
		case <-time.After(time.Second * time.Duration(timeout)):
			fmt.Println("The leader's election timeout, the node is changed to the follower status\n")
			rf.reDefault()
			return false
		case ok := <-rf.voteCh:
			if ok {
				rf.voteAdd()
				fmt.Printf("Get voting from other nodes, the current number of votes：%d\n", rf.vote)
			}
			if rf.vote > raftCount/2 && rf.currentLeader == "-1" {
				fmt.Println("Obtaining the number of votes that exceed one -half of the network node, this node has been elected to becomeleader")
				//Node status becomes 2, representing leader
				rf.setStatus(2)
				//The current leader is yourself
				rf.setCurrentLeader(rf.me)
				fmt.Println("Broadcast to other nodes...")
				go rf.broadcast("Raft.ConfirmationLeader", rf.node, func(ok bool) {
					fmt.Println(ok)
				})
				//Open the heartbeat detection channel
				rf.heartBeat <- true
				return true
			}
		}
	}
}

//Heartbeat detection method
func (rf *Raft) heartbeat() {
	//If YouReceive The Information On The Channel, You Will Perform A Fixed Frequency Heartbeat Detection To Other Nodes
	if <-rf.heartBeat {
		for {
			fmt.Println("This node starts to send heartbeat detection...")
			rf.broadcast("Raft.HeartbeatRe", rf.node, func(ok bool) {
				fmt.Println("reply received:", ok)
			})
			rf.lastHeartBeartTime = millisecond()
			time.Sleep(time.Second * time.Duration(heartBeatTimes))
		}
	}
}

//Random value
func randRange(min, max int64) int64 {
	//Time for heartbeat signal
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand.Int63n(max-min) + min
}

//Get the number of mills of time in the current time
func millisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//Setting term
func (rf *Raft) setTerm(term int) {
	rf.lock.Lock()
	rf.currentTerm = term
	rf.lock.Unlock()
}

//Who is set to vote
func (rf *Raft) setVoteFor(id string) {
	rf.lock.Lock()
	rf.votedFor = id
	rf.lock.Unlock()
}

//Set the current leader
func (rf *Raft) setCurrentLeader(leader string) {
	rf.lock.Lock()
	rf.currentLeader = leader
	rf.lock.Unlock()
}

//Set the current leader
func (rf *Raft) setStatus(state int) {
	rf.lock.Lock()
	rf.state = state
	rf.lock.Unlock()
}

//Voting accumulation
func (rf *Raft) voteAdd() {
	rf.lock.Lock()
	rf.vote++
	rf.lock.Unlock()
}

//Set the number of votes
func (rf *Raft) setVote(num int) {
	rf.lock.Lock()
	rf.vote = num
	rf.lock.Unlock()
}

//Restore the default settings
func (rf *Raft) reDefault() {
	rf.setVote(0)
	//rfCurrentLeader = "1"
	rf.setVoteFor("-1")
	rf.setStatus(0)
}
