# Twitch — Go Real Interview Questions & Deep Dives

> **Target Roles**: Real-Time Chat Infrastructure Engineer, Video Ingestion Engineer, Distributed Messaging Developer.

---

## 🏢 Why Twitch Asks Go
Twitch uses Go for its massive real-time chat architecture (IRC/WebSocket clusters handling tens of millions of concurrent chat connections). Twitch interviews focus on **Broadcaster Fan-Out Patterns**, **Sliding Window Chat Rate Limiters**, and **Memory-Efficient Client Session Management**.

---

## 📌 Question 1: High-Scale Chat Room Broadcast Engine

### Question
Write a high-performance Chat Room Broadcaster in Go that broadcasts messages from a chat channel publisher to $N$ connected subscriber goroutines without allowing a slow subscriber to block fast chat updates.

### Detailed Solution & Complete Code

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type Client struct {
	ID   int
	send chan string
}

type ChatRoom struct {
	mu      sync.RWMutex
	clients map[int]*Client
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients: make(map[int]*Client),
	}
}

func (r *ChatRoom) Register(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client.ID] = client
}

func (r *ChatRoom) Unregister(clientID int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.clients[clientID]; ok {
		close(c.send)
		delete(r.clients, clientID)
	}
}

func (r *ChatRoom) Broadcast(msg string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, client := range r.clients {
		select {
		case client.send <- msg:
		default:
			// Non-blocking send: Drop message for slow client to protect room throughput!
			fmt.Printf("Buffer full for client %d; dropping chat message!\n", client.ID)
		}
	}
}

func main() {
	room := NewChatRoom()

	c1 := &Client{ID: 1, send: make(chan string, 10)}
	c2 := &Client{ID: 2, send: make(chan string, 1)} // Small buffer to simulate slow subscriber

	room.Register(c1)
	room.Register(c2)

	// Receiver 1
	go func() {
		for msg := range c1.send {
			fmt.Println("Client 1 received:", msg)
		}
	}()

	// Broadcast flood
	for i := 1; i <= 3; i++ {
		room.Broadcast(fmt.Sprintf("PogChamp message %d", i))
	}

	time.Sleep(50 * time.Millisecond)
}
```
