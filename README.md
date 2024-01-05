# Raft Consensus Algorithm Demo in Go

This project is a demonstration of the Raft consensus algorithm implemented in Go. It's designed to help better understand the concept and workings of the algorithm.

## Features

This implementation includes the following features:

- Nodes are categorized into three roles: Leader, Follower, and Candidate.
- Elections are held among nodes to choose a Leader. At any given time, there can only be one Leader.
- The Leader node periodically sends heartbeat messages to Follower nodes.
- If a Follower node does not receive a heartbeat within a certain timeframe, it initiates a new election.
- Clients send messages to any node (say node A) via HTTP. If node A is not the Leader, the message is forwarded to the Leader node.
- Upon receiving a message from a client, the Leader broadcasts the message to all Follower nodes.
- Follower nodes, upon receiving the message, send an acknowledgment back to the Leader and wait for confirmation.
- Once the Leader receives acknowledgments from more than half of the nodes in the network, it logs the message locally and sends a confirmation of receipt to the Follower nodes.
- Follower nodes, upon receiving the confirmation, log the message.

## Running the Project

### Compilation

```go
go build -o raft
```

##### 2. Config File (the initial state)
You can customize the config file to define nodes' names and ports in the "nodes" field for the project. There are also additional options available for customization according to your requirements.


##### 3. nodes will randomly elect leaders (one of which is programmed to listen to HTTP requests), and the successful node will send a heartbeat to detect the other nodes.

##### 4. At this time, open a browser and use HTTP to access port 8080 of the local node. The message with the node needs to be printed simultaneously, such as:
`http://localhost:8080/REQ?Message=Oh, Hello Raft Algorithm`
You will see that the nodes are printing at the same time.

##### 5. Handling Leader Node Failure

If the leader node goes down, there is a timeout mechanism between the nodes. If a node does not receive a heartbeat for a certain period of time, it will automatically initiate a new election.

##### 6. Open the browser again and use HTTP to access port 8080 of the local node. The message with the node needs to be printed simultaneously. Can it still print simultaneously?
`http://localhost:8080/REQ?Message=Are you working?`

You will find that it can still print, because a new leader has been elected and the followers are functioning correctly. If more than half of the nodes on the network send feedback, the data can be printed.

##### 7. Restart the Failed Node

If you restart a failed node, it will receive a heartbeat from the leader node.

<hr>
