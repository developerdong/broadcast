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
		wg := sync.WaitGroup{}
		for m := range se {
			b.mu.RLock()
			wg.Add(len(b.receivers) - 1)
			for r := range b.receivers {
				if r != re {
					go func(r chan<- interface{}) {
						select {
						case r <- m:
						case <-time.After(timeout):
						}
						wg.Done()
					}(r)
				}
			}
			wg.Wait()
			b.mu.RUnlock()
		}
		b.mu.Lock()
		close(re)
		delete(b.receivers, re)
		b.mu.Unlock()
	}()
	return se, re
}
