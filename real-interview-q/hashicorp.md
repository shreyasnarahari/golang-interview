# HashiCorp — Go Real Interview Questions & Deep Dives

> **Target Roles**: Distributed Systems Engineer, Infrastructure Core Engineer (Terraform/Consul/Vault), Cloud Platform Engineer.

---

## 🏢 Why HashiCorp Asks Go
HashiCorp's entire flagship product line (Terraform, Vault, Consul, Nomad, Packer) is written in Go. HashiCorp interviews test **distributed state synchronization**, **Raft consensus state machines**, **`sync.Cond` broadcast coordination**, **plugin architectures over RPC**, and **context cancellation trees**.

---

## 📌 Question 1: State Machine Broadcast Synchronization with `sync.Cond`

### Question
In distributed systems like Consul or Raft, node state transitions (Leader $\rightarrow$ Follower $\rightarrow$ Candidate) must wake up all waiting goroutines simultaneously when state changes. Implement a thread-safe State Machine notifier using `sync.Cond`.

### Detailed Solution & Complete Code

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

func (s NodeState) String() string {
	return [...]string{"Follower", "Candidate", "Leader"}[s]
}

type ConsensusNode struct {
	mu    sync.Mutex
	cond  *sync.Cond
	state NodeState
}

func NewConsensusNode() *ConsensusNode {
	node := &ConsensusNode{state: Follower}
	node.cond = sync.NewCond(&node.mu)
	return node
}

func (n *ConsensusNode) SetState(newState NodeState) {
	n.mu.Lock()
	defer n.mu.Unlock()

	fmt.Printf("State transition: %s -> %s\n", n.state, newState)
	n.state = newState
	n.cond.Broadcast() // Wakes UP ALL waiting goroutines!
}

func (n *ConsensusNode) WaitForState(targetState NodeState) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for n.state != targetState {
		n.cond.Wait() // Releases lock, suspends goroutine, re-acquires lock upon wakeup
	}
	fmt.Printf("Goroutine woke up: state reached %s!\n", targetState)
}

func main() {
	node := NewConsensusNode()

	// Spawn multiple goroutines waiting for Leader state
	for i := 1; i <= 3; i++ {
		go func(id int) {
			node.WaitForState(Leader)
			fmt.Printf("Worker %d detected Leader elected!\n", id)
		}(i)
	}

	time.Sleep(50 * time.Millisecond)
	node.SetState(Candidate)

	time.Sleep(50 * time.Millisecond)
	node.SetState(Leader) // Broadcast triggers all 3 workers

	time.Sleep(50 * time.Millisecond)
}
```

---

## 📌 Question 2: In-Memory Key-Value Store with Replicating Commit Log

### Question
Write a simple concurrent in-memory Key-Value store featuring atomic transaction commits and an append-only commit log (similar to Raft/Consul WAL).

### Detailed Solution & Code

```go
package main

import (
	"fmt"
	"sync"
)

type CommandType int

const (
	OpSet CommandType = iota
	OpDelete
)

type LogEntry struct {
	Op    CommandType
	Key   string
	Value string
}

type ReplicatedKV struct {
	mu     sync.RWMutex
	store  map[string]string
	walLog []LogEntry
}

func NewReplicatedKV() *ReplicatedKV {
	return &ReplicatedKV{
		store: make(map[string]string),
	}
}

func (kv *ReplicatedKV) Apply(entry LogEntry) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.walLog = append(kv.walLog, entry)
	switch entry.Op {
	case OpSet:
		kv.store[entry.Key] = entry.Value
	case OpDelete:
		delete(kv.store, entry.Key)
	}
}

func (kv *ReplicatedKV) Get(key string) (string, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	val, ok := kv.store[key]
	return val, ok
}

func main() {
	kv := NewReplicatedKV()
	kv.Apply(LogEntry{Op: OpSet, Key: "consul/datacenter", Value: "dc1"})
	kv.Apply(LogEntry{Op: OpSet, Key: "consul/retry_join", Value: "10.0.0.1"})

	val, _ := kv.Get("consul/datacenter")
	fmt.Println("Datacenter:", val)
}
```
