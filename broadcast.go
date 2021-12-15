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
		beforeGroup, afterGroup := make([]chan<- interface{}, 0), make([]chan<- interface{}, 0)
		for m := range se {
			// reset the timer
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timeout)
			// broadcast messages
			beforeGroup, afterGroup = beforeGroup[:0], afterGroup[:0]
			b.mu.RLock()
			for r := range b.receivers {
				if r != re {
					select {
					case r <- m:
					case <-timer.C:
						goto finish
					default:
						beforeGroup = append(beforeGroup, r)
					}
				}
			}
			for {
				if len(beforeGroup) == 0 {
					goto finish
				}
				for _, r := range beforeGroup {
					select {
					case r <- m:
					case <-timer.C:
						goto finish
					default:
						afterGroup = append(afterGroup, r)
					}
				}
				beforeGroup, afterGroup = afterGroup, afterGroup[:0]
			}
		finish:
			b.mu.RUnlock()
		}
		b.mu.Lock()
		close(re)
		delete(b.receivers, re)
		b.mu.Unlock()
	}()
	return se, re
}
