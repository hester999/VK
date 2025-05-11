package main

import (
	"VK/internal/config"
	"VK/internal/grpc/subpub_service"
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	serverAddr := flag.String("server", cfg.Client.DefaultServer, "адрес сервера")
	flag.Parse()

	client, err := subpub_service.NewClient(*serverAddr) // Создаем новый клиент
	if err != nil {
		slog.Error("failed to create client", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background()) 
	defer cancel()

	for _, topic := range cfg.Client.Topics {
		err := client.Publish(ctx, topic, "init")
		if err != nil {
			slog.Error("failed to create topic", "topic", topic, "error", err)
			continue
		}
		slog.Info("created topic", "topic", topic)
		time.Sleep(100 * time.Millisecond)
	}


	messageChans := make(map[string]<-chan string) 
	for _, topic := range cfg.Client.Topics {
		messageChan, err := client.Subscribe(ctx, topic)
		if err != nil {
			slog.Error("failed to subscribe", "topic", topic, "error", err)
			continue
		}
		messageChans[topic] = messageChan
		slog.Info("subscribed to topic", "topic", topic)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)


	messageDone := make(chan struct{})
	go func() {
		defer close(messageDone)
		for topic, ch := range messageChans {
			go func(topic string, ch <-chan string) {
				for msg := range ch {
					if msg != "init" {
						slog.Info("received message", "topic", topic, "message", msg)
					}
				}
			}(topic, ch)
		}
		<-ctx.Done()
	}()


	publishDone := make(chan struct{})
	go func() {
		defer close(publishDone)
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
					slog.Error("failed to publish", "topic", topic, "error", err)
					continue
				}
				slog.Info("published message", "topic", topic, "message", msg)
				time.Sleep(500 * time.Millisecond)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	<-stop
	slog.Info("shutting down client...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	select {
	case <-shutdownCtx.Done():
		slog.Warn("shutdown timeout")
	case <-messageDone:
		slog.Info("message handler stopped")
	case <-publishDone:
		slog.Info("publisher stopped")
	}

	slog.Info("client stopped")
}
