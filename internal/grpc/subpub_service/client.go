package subpub_service

import (
	pubsub "VK/api/proto"
	"VK/internal/pkg/logger"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"VK/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client pubsub.PubSubClient
	logger *slog.Logger
}
// Создаем новый клиент
func NewClient(serverAddress string) (*Client, error) {
	conn, err := grpc.Dial(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: pubsub.NewPubSubClient(conn),
		logger: logger.New(),
	}, nil
}
// Подписываемся на топик
func (c *Client) Subscribe(ctx context.Context, topic string) (<-chan string, error) {
	stream, err := c.client.Subscribe(ctx, &pubsub.SubscribeRequest{
		Key: topic,
	})
	if err != nil {
		return nil, err
	}

	messageChan := make(chan string)

	go func() {
		defer close(messageChan)
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				c.logger.Info("stream closed", "topic", topic)
				return
			}
			if err != nil {
				c.logger.Error("error receiving message", "topic", topic, "error", err)
				return
			}
			messageChan <- event.Data
		}
	}()

	return messageChan, nil
}
// Публикуем сообщение в топик
func (c *Client) Publish(ctx context.Context, topic, message string) error {
	_, err := c.client.Publish(ctx, &pubsub.PublishRequest{
		Key:  topic,
		Data: message,
	})
	return err
}

func ExampleUsage() {
	client, err := NewClient("localhost:50051")
	if err != nil {
		panic(err)
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Сначала создаем все топики
	for _, topic := range cfg.Client.Topics {
		// Создаем топик через публикацию nil
		err := client.Publish(ctx, topic, "init")
		if err != nil {
			fmt.Printf("Error creating topic %s: %v\n", topic, err)
			continue
		}
		fmt.Printf("Created topic: %s\n", topic)
		time.Sleep(100 * time.Millisecond) 
	}

	// Подписываемся на все топики из конфигурации
	for _, topic := range cfg.Client.Topics {
		messageChan, err := client.Subscribe(ctx, topic)
		if err != nil {
			fmt.Printf("Error subscribing to topic %s: %v\n", topic, err)
			continue
		}

		// Запускаем горутину для каждого топика
		go func(topic string, ch <-chan string) {
			for msg := range ch {
				if msg != "init" { 
					fmt.Printf("[%s] Received: %s\n", topic, msg)
				}
			}
		}(topic, messageChan)
	}

	// Отправляем сообщения в разные топики
	messages := map[string][]string{
		"news":    {"Breaking news: Important event", "Latest updates from the world"},
		"weather": {"Sunny day ahead", "Rain expected tomorrow"},
		"sports":  {"Team wins championship", "New record set"},
		"tech":    {"New gadget released", "Software update available"},
		"music":   {"New album released", "Concert announced"},
	}

	
	for i := 0; i < 2; i++ { 
		for topic, msgs := range messages {
			msg := msgs[i]
			err := client.Publish(ctx, topic, msg)
			if err != nil {
				fmt.Printf("Error publishing to topic %s: %v\n", topic, err)
				continue
			}
			fmt.Printf("Published to [%s]: %s\n", topic, msg)
			time.Sleep(500 * time.Millisecond) 
		}
		time.Sleep(1 * time.Second) 
	}
}
