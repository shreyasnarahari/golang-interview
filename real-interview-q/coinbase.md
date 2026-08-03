# Coinbase — Go Real Interview Questions & Deep Dives

> **Target Roles**: Crypto Infrastructure Engineer, Order Matching Engine Developer, Blockchain Indexer Developer.

---

## 🏢 Why Coinbase Asks Go
Coinbase uses Go for its high-throughput crypto trading engine, wallet infrastructure, and node indexing pipelines. Coinbase interviews test **Order Book Data Structures**, **Goroutine Channel Multiplexing for WebSockets**, and **Thread-Safe Numeric State Machines**.

---

## 📌 Question 1: Concurrent Limit Order Book Matching Engine

### Question
Design a simplified in-memory Limit Order Book in Go supporting concurrent buy/sell limit order matching using doubly linked lists and `sync.RWMutex`.

### Detailed Solution & Complete Code

```go
package main

import (
	"fmt"
	"sync"
)

type OrderType int

const (
	Buy OrderType = iota
	Sell
)

type Order struct {
	ID    int64
	Type  OrderType
	Price float64
	Amount float64
}

type OrderBook struct {
	mu   sync.Mutex
	buys []*Order // Sorted descending by price
	sells []*Order // Sorted ascending by price
}

func NewOrderBook() *OrderBook {
	return &OrderBook{}
}

func (ob *OrderBook) PlaceOrder(order *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if order.Type == Buy {
		// Match against sells
		for len(ob.sells) > 0 && ob.sells[0].Price <= order.Price && order.Amount > 0 {
			match := ob.sells[0]
			tradeAmount := order.Amount
			if match.Amount < tradeAmount {
				tradeAmount = match.Amount
			}

			fmt.Printf("MATCH EXECUTED: Buy %d and Sell %d @ $%.2f (Qty: %.2f)\n",
				order.ID, match.ID, match.Price, tradeAmount)

			order.Amount -= tradeAmount
			match.Amount -= tradeAmount

			if match.Amount == 0 {
				ob.sells = ob.sells[1:]
			}
		}

		if order.Amount > 0 {
			ob.buys = append(ob.buys, order)
		}
	} else {
		// Match against buys
		for len(ob.buys) > 0 && ob.buys[0].Price >= order.Price && order.Amount > 0 {
			match := ob.buys[0]
			tradeAmount := order.Amount
			if match.Amount < tradeAmount {
				tradeAmount = match.Amount
			}

			fmt.Printf("MATCH EXECUTED: Sell %d and Buy %d @ $%.2f (Qty: %.2f)\n",
				order.ID, match.ID, match.Price, tradeAmount)

			order.Amount -= tradeAmount
			match.Amount -= tradeAmount

			if match.Amount == 0 {
				ob.buys = ob.buys[1:]
			}
		}

		if order.Amount > 0 {
			ob.sells = append(ob.sells, order)
		}
	}
}

func main() {
	book := NewOrderBook()
	book.PlaceOrder(&Order{ID: 1, Type: Sell, Price: 30000.0, Amount: 1.5})
	book.PlaceOrder(&Order{ID: 2, Type: Buy, Price: 30500.0, Amount: 1.0}) // Matches 1.0 BTC @ $30000
}
```
