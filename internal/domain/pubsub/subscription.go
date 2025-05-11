package pubsub

import (
	"log/slog"
	"VK/internal/entity"
)

type Subscription interface {
	Unsubscribe()
}

type subscription struct { // Структура для хранения подписки
	pub     *SubPubImpl // Публикатор
	subject string // Топик
	sub     *entity.Subscriber // Подписчик
}

func (s *subscription) Unsubscribe() { // Метод для отписки от топика
	s.pub.mu.Lock()
	defer s.pub.mu.Unlock()

	subs := s.pub.subs[s.subject] // Получаем все подписчики на топик

	for i, del := range subs {
		if del == s.sub { 
			s.pub.subs[s.subject] = append(subs[:i], subs[i+1:]...) 	
			s.pub.logger.Debug("Unsubscribed", slog.String("subject", s.subject))
			close(del.Ch)
			break
		}
	}
}
