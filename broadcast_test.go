package broadcast

import (
	"sync"
	"testing"
	"time"
)

func TestBroadCaster_Join(t *testing.T) {
	const N = 10
	broadcaster := New()
	wg := sync.WaitGroup{}
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			sender, receiver := broadcaster.Join(time.Second)
			go func() {
				sender <- i
				close(sender)
			}()
			for m := range receiver {
				if m.(int) == i {
					t.Errorf("receiver %d: received a same number %v", i, m)
					return
				}
				t.Logf("receiver %d: %v", i, m)
			}
		}(i)
	}
	wg.Wait()
}

func TestBroadCaster_Join2(t *testing.T) {
	const N = 10
	broadcaster := New()
	wg := sync.WaitGroup{}
	wg.Add(N)
	for i := 0; i < N; i++ {
		sender, receiver := broadcaster.Join(time.Second)
		go func(i int) {
			defer wg.Done()
			received := make(map[interface{}]struct{})
			for m := range receiver {
				if _, ok := received[m]; ok {
					t.Errorf("receiver %d: received %v twice", i, m)
					return
				}
				received[m] = struct{}{}
				t.Logf("receiver %d: %v", i, m)
				if len(received) == N {
					close(sender)
				}
			}
		}(i)
	}
	sender, _ := broadcaster.Join(time.Second)
	for i := 0; i < N; i++ {
		sender <- i
	}
	close(sender)
	wg.Wait()
}
