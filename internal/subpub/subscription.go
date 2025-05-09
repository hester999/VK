package subpub

import "VK/internal/entity"

type Subscription interface {
	Unsubscribe()
}

type subscription struct {
	pub     *SubPubImpl
	subject string
	sub     *entity.Subscriber
}

func (s *subscription) Unsubscribe() {
	s.pub.mu.Lock()
	defer s.pub.mu.Unlock()

	subs := s.pub.subs[s.subject]

	for i, del := range subs {

		if del == s.sub {
			s.pub.subs[s.subject] = append(subs[:i], subs[i+1:]...)

			close(del.Ch)
			break
		}
	}

}
