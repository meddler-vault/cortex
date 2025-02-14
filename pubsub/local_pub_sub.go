package pubsub

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Subscriber struct
type Subscriber struct {
	id     int
	filter string
	ch     chan string
}

// PubSub struct
type PubSub struct {
	identifier  string
	mu          sync.RWMutex
	subscribers map[int]*Subscriber
	nextID      int
}

// NewPubSub creates a new PubSub instance
func NewPubSub(identifier string) *PubSub {
	return &PubSub{
		identifier:  identifier,
		subscribers: make(map[int]*Subscriber),
	}
}

// Subscribe adds a new subscriber with a given filter and returns the channel & ID
func (ps *PubSub) Subscribe(filter string) (int, <-chan string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	id := ps.nextID
	ps.nextID++
	sub := &Subscriber{
		id:     id,
		filter: filter,
		ch:     make(chan string, 1), // Buffered to avoid blocking
	}
	ps.subscribers[id] = sub
	log.Println("local_pubsub_subscription", filter, id)
	return id, sub.ch
}

// Unsubscribe removes a subscriber by ID and closes their channel
func (ps *PubSub) Unsubscribe(id int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if sub, ok := ps.subscribers[id]; ok {
		close(sub.ch) // Close channel to notify the subscriber
		delete(ps.subscribers, id)
	}
}

// Publish sends a message to all matching subscribers
func (ps *PubSub) Publish(msg string) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	for _, sub := range ps.subscribers {
		if sub.filter == msg {
			select {
			case sub.ch <- msg:
			default: // Avoid blocking if the channel is full
			}
		}
	}
}

func main() {

	ps := NewPubSub("test")
	// Simulate publishing messages
	go func() {
		words := []string{"hello", "foo", "world", "bar", "hello", "world"}
		for _, word := range words {
			fmt.Println("Publishing:", word)
			ps.Publish(word)
			time.Sleep(1 * time.Second)
		}
	}()
	// Subscriber 1 (looking for "hello")
	id1, ch1 := ps.Subscribe("hello")
	go func() {
		for msg := range ch1 {
			fmt.Println("Subscriber 1 received:", msg)
		}
		fmt.Println("Subscriber 1 unsubscribed and exited.")
	}()

	// Subscriber 2 (looking for "world")
	id2, ch2 := ps.Subscribe("world")
	go func() {
		for msg := range ch2 {
			fmt.Println("Subscriber 2 received:", msg)
		}
		fmt.Println("Subscriber 2 unsubscribed and exited.")
	}()

	// Unsubscribe after 3 seconds
	time.Sleep(3 * time.Second)
	fmt.Println("Unsubscribing Subscriber 1")
	ps.Unsubscribe(id1)

	// Unsubscribe another subscriber after 5 seconds
	time.Sleep(2 * time.Second)
	fmt.Println("Unsubscribing Subscriber 2")
	ps.Unsubscribe(id2)

	// Allow time for everything to complete
	time.Sleep(2 * time.Second)
}
