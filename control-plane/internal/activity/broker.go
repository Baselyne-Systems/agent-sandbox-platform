package activity

import (
	"sync"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

const subscriberBufferSize = 256

// Broker provides in-process pub/sub for action records.
// Subscribers receive a buffered channel; slow consumers may miss events.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[uint64]chan *models.ActionRecord
	nextID      uint64
}

// NewBroker creates a new action record broker.
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[uint64]chan *models.ActionRecord),
	}
}

// Subscribe returns a channel that receives action records and an ID for unsubscribing.
func (b *Broker) Subscribe() (uint64, <-chan *models.ActionRecord) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	id := b.nextID
	ch := make(chan *models.ActionRecord, subscriberBufferSize)
	b.subscribers[id] = ch
	return id, ch
}

// Unsubscribe removes a subscriber and closes its channel.
func (b *Broker) Unsubscribe(id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscribers[id]; ok {
		close(ch)
		delete(b.subscribers, id)
	}
}

// Publish sends an action record to all subscribers.
// Non-blocking: if a subscriber's buffer is full, the event is dropped for that subscriber.
func (b *Broker) Publish(record *models.ActionRecord) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers {
		select {
		case ch <- record:
		default:
			// Subscriber is slow, drop the event.
		}
	}
}
