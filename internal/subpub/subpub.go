package subpub

import (
	"VK/internal/entity"
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	SubjectNotExist = errors.New("subject not exist")
)

type SubPub interface {
	Subscribe(subject string, cb entity.MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{})
	Close(ctx context.Context) error
}

type SubPubImpl struct {
	mu     sync.Mutex
	subs   map[string][]*entity.Subscriber
	wg     sync.WaitGroup
	closed bool
}

func (s *SubPubImpl) Subscribe(subject string, cb entity.MessageHandler) (Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("pubsub is closed")
	}

	if _, ok := s.subs[subject]; !ok {
		return nil, fmt.Errorf("subject %s not exist: %w", subject, SubjectNotExist)
	}

	sub := &entity.Subscriber{
		Ch: make(chan interface{}, 10),
		Cb: cb,
	}
	s.subs[subject] = append(s.subs[subject], sub)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for msg := range sub.Ch {
			sub.Cb(msg)
		}
	}()

	return &subscription{
		pub:     s,
		subject: subject,
		sub:     sub,
	}, nil
}

func (s *SubPubImpl) Publish(subject string, msg interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	if _, ok := s.subs[subject]; !ok {
		s.subs[subject] = []*entity.Subscriber{}
	}

	for _, sub := range s.subs[subject] {
		select {
		case sub.Ch <- msg:
		default:

		}
	}
}

func (s *SubPubImpl) Close(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true

	for _, subs := range s.subs {
		for _, sub := range subs {
			close(sub.Ch)
		}
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func NewSubPub() SubPub {
	return &SubPubImpl{
		subs: make(map[string][]*entity.Subscriber),
	}
}
