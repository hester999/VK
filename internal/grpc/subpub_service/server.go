package subpub_service

import (
	pubsub "VK/api/proto"
	domainpubsub "VK/internal/domain/pubsub"
	"VK/internal/pkg/logger"
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serverAPI struct {
	subpub domainpubsub.SubPub 
	pubsub.UnimplementedPubSubServer
	logger *slog.Logger
}
// Регистрируем сервер
func Register(grpcServer *grpc.Server, ps domainpubsub.SubPub) {
	pubsub.RegisterPubSubServer(grpcServer, &serverAPI{
		subpub: ps, 
		logger: logger.New(),
	})
}

func (s *serverAPI) Subscribe(
	req *pubsub.SubscribeRequest,
	stream pubsub.PubSub_SubscribeServer,
) error {
	topic := req.GetKey() // Получаем топик из запроса
	if topic == "" {  // Если топик пустой, то возвращаем ошибку
		s.logger.Error("subscribe request failed", "error", "topic cannot be empty")
		return status.Errorf(codes.InvalidArgument, "topic cannot be empty")
	}

	s.logger.Info("new subscription request", "topic", topic)
	messageChan := make(chan string, 10)
 
	sub, err := s.subpub.Subscribe(topic, func(msg any) { 
		if _, ok := msg.(string); !ok {
			s.logger.Warn("received non-string message", "topic", topic)
			return
		}
		messageChan <- msg.(string)
	})
	if errors.Is(err, domainpubsub.ErrSubjectNotExist) {
		s.logger.Error("subscribe failed", "topic", topic, "error", err)
		return status.Errorf(codes.NotFound, "subject %s not exist: %v", topic, err)
	}

	if err != nil {
		s.logger.Error("subscribe failed", "topic", topic, "error", err)
		return status.Errorf(codes.Internal, "failed to subscribe: %v", err)
	}

	s.logger.Info("subscription established", "topic", topic)

	for {
		select { 
		case msg := <-messageChan:
			event := &pubsub.Event{
				Data: msg,
			}
			if err := stream.Send(event); err != nil {
				s.logger.Error("failed to send message", "topic", topic, "error", err)
				return status.Errorf(codes.Internal, "failed to send message: %v", err)
			}
			s.logger.Debug("message sent to subscriber", "topic", topic)
		case <-stream.Context().Done():
			s.logger.Info("subscription closed", "topic", topic)
			sub.Unsubscribe()
			return nil
		}
	}
}


func (s *serverAPI) Publish(
	ctx context.Context, 
	req *pubsub.PublishRequest, 
) (*emptypb.Empty, error) {
	topic := req.GetKey()
	if topic == "" {
		s.logger.Error("publish request failed", "error", "topic cannot be empty")
		return nil, status.Errorf(codes.InvalidArgument, "topic cannot be empty")
	}
	data := req.GetData()
	if data == "" {
		s.logger.Error("publish request failed", "error", "data cannot be empty")
		return nil, status.Errorf(codes.InvalidArgument, "data cannot be empty")
	}

	s.logger.Info("publishing message", "topic", topic)
	s.subpub.Publish(topic, data)
	s.logger.Debug("message published", "topic", topic)

	return &emptypb.Empty{}, nil
}
