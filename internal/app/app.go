package app

import (
	"VK/internal/config"
	"VK/internal/domain/pubsub"
	"VK/internal/grpc/subpub_service"
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

type App struct {
	GRPCServer *grpc.Server
	PubSub     pubsub.SubPub
	Logger     *slog.Logger
}

func NewApp(cfg *config.Config, logger *slog.Logger) *App {

	pubsub := pubsub.NewSubPub(logger)
	grpcServer := grpc.NewServer()
	subpub_service.Register(grpcServer, pubsub)
	return &App{GRPCServer: grpcServer, PubSub: pubsub, Logger: logger}
}

// Функция запуска сервера,  тут пытался сделать graceful shutdown
func (a *App) Run(addr, port string) error {
	lis, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		a.Logger.Error("failed to listen", slog.String("err", err.Error()))
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		a.Logger.Info("gRPC server started", slog.String("addr:port", addr+":"+port))
		if err := a.GRPCServer.Serve(lis); err != nil {
			a.Logger.Error("failed to serve", slog.String("err", err.Error()))
		}
	}()

	<-stop
	a.Logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		a.GRPCServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		a.Logger.Warn("shutdown timeout, forcing stop")
		a.GRPCServer.Stop()
	case <-stopped:
		a.Logger.Info("server stopped gracefully")
	}

	return nil
}
