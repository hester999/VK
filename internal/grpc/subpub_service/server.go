package subpub_service

import (
	pubsub "VK/internal/gen/go/pubsub"
	"VK/internal/subpub"
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serverAPI struct {
	subpub subpub.SubPub
	pubsub.UnimplementedPubSubServer
}

func Register(grpcServer *grpc.Server, ps subpub.SubPub) {
	pubsub.RegisterPubSubServer(grpcServer, &serverAPI{
		subpub: ps,
	})
}


func (s *serverAPI) Subscribe(
	req *pubsub.SubscribeRequest, 
	stream pubsub.PubSub_SubscribeServer, 
) error {
	
	topic := req.GetKey()
	if topic == "" {
		return status.Errorf(codes.InvalidArgument, "topic cannot be empty")
	}

	messageChan := make(chan string, 10)

	
	sub, err := s.subpub.Subscribe(topic, func(msg any) {
		if _, ok := msg.(string); !ok {
			return
		}
		messageChan <- msg.(string)
	})
	if errors.Is(err, subpub.SubjectNotExist) {
		return status.Errorf(codes.NotFound, "subject %s not exist: %v", topic, err)
	}

	if err != nil {
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	
	for {
		select {
		case msg := <-messageChan:
			event := &pubsub.Event{
				Data: msg,
			}
			if err := stream.Send(event); err != nil {
				return status.Errorf(codes.Internal, "failed to send message: %v", err)
			}
		case <-stream.Context().Done():
			sub.Unsubscribe()
			return nil
		}
	}

}

// Реализация метода Publish для gRPC PubSub сервиса.
// Этот метод должен принимать запрос на публикацию и рассылать событие всем подписчикам.
func (s *serverAPI) Publish(
	ctx context.Context, // Контекст запроса
	req *pubsub.PublishRequest, // Запрос с ключом (темой) и данными для публикации
) (*emptypb.Empty, error) {
	// 1. Получить ключ (тему) и данные из запроса
	topic := req.GetKey()
	if topic == "" {
		return nil, status.Errorf(codes.InvalidArgument, "topic cannot be empty")
	}
	data := req.GetData()
	if data == "" {
		return nil, status.Errorf(codes.InvalidArgument, "data cannot be empty")
	}
	
	// 2. Вызвать s.subpub.Publish для рассылки сообщения
	s.subpub.Publish(topic, data)
	
	// 3. Вернуть пустой ответ или ошибку
	return &emptypb.Empty{}, nil
}
