package subpub_tests

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"VK/internal/domain/pubsub"
	"VK/internal/pkg/logger"
)
// Тесты для проверки функциональности подписки и публикации
func TestSubscribeAndPublish(t *testing.T) {
	logger := logger.New()
	bus := pubsub.NewSubPub(logger)
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
// Тест для проверки отписки от топика
func TestUnsubscribe(t *testing.T) {
	logger := logger.New()
	bus := pubsub.NewSubPub(logger)
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
// Тест для проверки закрытия всех подписок
func TestCloseStopsAll(t *testing.T) {
	logger := logger.New()
	bus := pubsub.NewSubPub(logger)
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
// Тест для проверки подписки на несуществующий топик
func TestSubscribeToUnknownTopicFails(t *testing.T) {
	logger := logger.New()
	bus := pubsub.NewSubPub(logger)

	_, err := bus.Subscribe("nonexistent", func(msg interface{}) {})
	if err == nil {
		t.Fatal("expected error when subscribing to unknown topic")
	}

	if !errors.Is(err, pubsub.ErrSubjectNotExist) {
		t.Fatalf("expected ErrSubjectNotExist error, got: %v", err)
	}
}
// Тест для проверки закрытия с таймаутом контекста
func TestCloseWithTimeoutContext(t *testing.T) {
	logger := logger.New()
	bus := pubsub.NewSubPub(logger)
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
// Тест для проверки подписки после закрытия
func TestSubscribeAfterCloseFails(t *testing.T) {
	logger := logger.New()
	bus := pubsub.NewSubPub(logger)
	topic := "x"
	bus.Publish(topic, nil)

	_ = bus.Close(context.Background()) 

	_, err := bus.Subscribe(topic, func(msg interface{}) {})
	if err == nil {
		t.Fatal("expected error when subscribing after close")
	}
}
// Тест для проверки публикации после закрытия
func TestPublishAfterCloseDoesNothing(t *testing.T) {
	logger := logger.New()
	bus := pubsub.NewSubPub(logger)
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

