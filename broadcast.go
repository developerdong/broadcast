package broadcast

import (
	"sync"
	"time"
)

type BroadCaster struct {
	mu        sync.RWMutex
	receivers map[chan<- interface{}]struct{}
}

func New() *BroadCaster {
	return &BroadCaster{
		mu:        sync.RWMutex{},
		receivers: make(map[chan<- interface{}]struct{}),
	}
}

// Join returns a sender and a receiver, which are for sending messages to and
// receiving messages from respectively. The parameter timeout is to not block
// the sending action to other clients if one client is slow or unresponsive. If
// you want to exit the broadcast group, you must close the returned sender
// channel.
func (b *BroadCaster) Join(timeout time.Duration) (sender chan<- interface{}, receiver <-chan interface{}) {
	se, re := make(chan interface{}), make(chan interface{})
	b.mu.Lock()
	b.receivers[re] = struct{}{}
	b.mu.Unlock()
	go func() {
		timer := time.NewTimer(timeout)
		for m := range se {
			// reset the timer
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeout)
			// broadcast messages
			blocking := true
			b.mu.RLock()
			for r := range b.receivers {
				if r != re {
					if blocking {
						select {
						case r <- m:
						case <-timer.C:
							// If this operation has expired, the next operations should be non-blocking.
							blocking = false
						}
					} else {
						select {
						case r <- m:
						default:
						}
					}
				}
			}
			b.mu.RUnlock()
		}
		b.mu.Lock()
		close(re)
		delete(b.receivers, re)
		b.mu.Unlock()
	}()
	return se, re
}
