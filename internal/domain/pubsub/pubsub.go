package pubsub

import (
	"VK/internal/entity"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)
// Ошибка для случая, когда топик не существует
var (
	ErrSubjectNotExist = errors.New("subject not exist")
)
// Интерфейс для подписки и публикации
type SubPub interface {
	Subscribe(subject string, cb entity.MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{})
	Close(ctx context.Context) error
}
// Реализация интерфейса для подписки и публикации
type SubPubImpl struct {
	mu     sync.Mutex // Мьютекс для синхронизации доступа к подпискам
	subs   map[string][]*entity.Subscriber // Мапа для хранения подписок
	wg     sync.WaitGroup // Группа ожидания для ожидания завершения всех горутин
	closed bool // Флаг для проверки закрытия подписки
	logger *slog.Logger // Логгер для логирования
}
// Функция для подписки на топик
func (s *SubPubImpl) Subscribe(subject string, cb entity.MessageHandler) (Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		s.logger.Warn("Subscribe called after close", slog.String("subject", subject))
		return nil, errors.New("pubsub is closed")
	}

	if _, ok := s.subs[subject]; !ok {
		s.logger.Warn("Subject not exist", slog.String("subject", subject))
		return nil, fmt.Errorf("subject %s not exist: %w", subject, ErrSubjectNotExist)
	}

	// Создание нового подписчика
	sub := &entity.Subscriber{
		Ch: make(chan interface{}, 10),
		Cb: cb, // Обработчик сообщений
	}
	s.subs[subject] = append(s.subs[subject], sub)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for msg := range sub.Ch {
			s.logger.Debug("Message received", slog.String("subject", subject))
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

	if s.closed { // Если подписка закрыта, то нельзя публиковать сообщения
		s.logger.Warn("Publish called after close", slog.String("subject", subject))
		return
	}

	if _, ok := s.subs[subject]; !ok { // Если топик не существует, то создаем новый
		s.subs[subject] = []*entity.Subscriber{}
		s.logger.Info("Created new subject", slog.String("subject", subject))
	}

	for _, sub := range s.subs[subject] {
		select {
		case sub.Ch <- msg: // Отправляем сообщение в канал подписчика
			s.logger.Debug("Message published", slog.String("subject", subject))
		default:
			s.logger.Warn("Subscriber channel full", slog.String("subject", subject))
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

func NewSubPub(logger *slog.Logger) SubPub {
	return &SubPubImpl{
		subs:   make(map[string][]*entity.Subscriber),
		logger: logger,
	}
}
