package broadcast

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestBroadCaster_Back ensures that a message will not be received by the sender itself.
func TestBroadCaster_Back(t *testing.T) {
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

// TestBroadCaster_Duplicate ensures that a receiver will not receive a message twice.
func TestBroadCaster_Duplicate(t *testing.T) {
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

func TestBroadCaster_ZeroTimeout(t *testing.T) {
	const N = 10000
	const Timeout = 0
	broadcaster := New()
	sender, _ := broadcaster.Join(Timeout)
	_, receiver := broadcaster.Join(Timeout)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer cancel()
		for i := 0; i < N; i++ {
			select {
			case <-ctx.Done():
				return
			case sender <- struct{}{}:
			}
		}
	}()
	go func() {
		defer wg.Done()
		defer cancel()
		for i := 0; i < N; i++ {
			select {
			case <-ctx.Done():
				return
			case <-receiver:
			}
		}
	}()
	wg.Wait()
}

func TestBroadCaster_Block(t *testing.T) {
	broadcaster := New()
	sender, _ := broadcaster.Join(time.Second)
	_, _ = broadcaster.Join(time.Second)
	startTime := time.Now()
	sender <- struct{}{}
	sender <- struct{}{}
	if duration := time.Since(startTime); duration < time.Second {
		t.Errorf("the operation should be blocked at least 1 second, but we got %v", duration)
	}
}

func BenchmarkBroadCaster(b *testing.B) {
	broadcaster := New()
	sender, _ := broadcaster.Join(time.Second)
	for i := 0; i < 1024; i++ {
		go func() {
			_, receiver := broadcaster.Join(time.Second)
			for range receiver {
			}
		}()
	}
	for i := 0; i < b.N; i++ {
		sender <- i
	}
}
