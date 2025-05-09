package subpub_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	_ "VK/internal/entity"
	"VK/internal/subpub"
)

func TestSubscribeAndPublish(t *testing.T) {
	bus := subpub.NewSubPub()
	topic := "cats"
	bus.Publish(topic, nil)

	var mu sync.Mutex
	var received []interface{}

	handler := func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg)
	}

	sub, err := bus.Subscribe(topic, handler)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	defer sub.Unsubscribe()

	bus.Publish(topic, "meow")
	bus.Publish(topic, "purr")

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(received))
	}
	if received[0] != "meow" || received[1] != "purr" {
		t.Errorf("unexpected message order: %v", received)
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := subpub.NewSubPub()
	topic := "dogs"
	bus.Publish(topic, nil)

	var mu sync.Mutex
	var received []interface{}

	handler := func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg)
	}

	sub, err := bus.Subscribe(topic, handler)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	bus.Publish(topic, "woof")
	time.Sleep(50 * time.Millisecond)

	sub.Unsubscribe()

	bus.Publish(topic, "bark")
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 || received[0] != "woof" {
		t.Errorf("expected only 'woof', got: %v", received)
	}
}

func TestCloseStopsAll(t *testing.T) {
	bus := subpub.NewSubPub()
	topic := "birds"
	bus.Publish(topic, nil)

	received := make(chan interface{}, 10)

	handler := func(msg interface{}) {
		received <- msg
	}

	_, err := bus.Subscribe(topic, handler)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	bus.Publish(topic, "chirp")
	bus.Publish(topic, "tweet")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = bus.Close(ctx)
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	select {
	case msg := <-received:
		if msg != "chirp" && msg != "tweet" {
			t.Errorf("unexpected message: %v", msg)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("expected message after close, got nothing")
	}
}

func TestSubscribeToUnknownTopicFails(t *testing.T) {
	bus := subpub.NewSubPub()

	_, err := bus.Subscribe("nonexistent", func(msg interface{}) {})
	if err == nil {
		t.Fatal("expected error when subscribing to unknown topic")
	}

	if !errors.Is(err, subpub.SubjectNotExist) {
		t.Fatalf("expected SubjectNotExist error, got: %v", err)
	}
}

func TestCloseWithTimeoutContext(t *testing.T) {
	bus := subpub.NewSubPub()
	topic := "slow"
	bus.Publish(topic, nil)

	handler := func(msg interface{}) {
		time.Sleep(2 * time.Second)
	}

	_, _ = bus.Subscribe(topic, handler)

	bus.Publish(topic, "delayed")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := bus.Close(ctx)
	if err == nil {
		t.Fatal("expected error due to context timeout")
	}
}

func TestSubscribeAfterCloseFails(t *testing.T) {
	bus := subpub.NewSubPub()
	topic := "x"
	bus.Publish(topic, nil)

	_ = bus.Close(context.Background()) // закрываем

	_, err := bus.Subscribe(topic, func(msg interface{}) {})
	if err == nil {
		t.Fatal("expected error when subscribing after close")
	}
}

func TestPublishAfterCloseDoesNothing(t *testing.T) {
	bus := subpub.NewSubPub()
	topic := "y"
	bus.Publish(topic, nil)

	received := make(chan interface{}, 1)
	_, _ = bus.Subscribe(topic, func(msg interface{}) {
		received <- msg
	})

	_ = bus.Close(context.Background())
	bus.Publish(topic, "should not arrive")

	time.Sleep(50 * time.Millisecond)

	select {
	case <-received:
		t.Fatal("unexpected message received after close")
	default:

	}
}

func TestCloseIsIdempotent(t *testing.T) {
	bus := subpub.NewSubPub()
	topic := "z"
	bus.Publish(topic, nil)

	_, _ = bus.Subscribe(topic, func(msg interface{}) {})

	ctx := context.Background()
	if err := bus.Close(ctx); err != nil {
		t.Fatalf("first close failed: %v", err)
	}

	if err := bus.Close(ctx); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
}
