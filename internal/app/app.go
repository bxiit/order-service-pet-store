package app

import (
	grpcapp "github.com/bxiit/order-service-pet-store/internal/app/grpc"
	"github.com/bxiit/order-service-pet-store/internal/data"
	"github.com/bxiit/order-service-pet-store/internal/services/order"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dsn string,
	tokenTTL time.Duration,
) *App {
	storage, err := data.New(dsn)
	if err != nil {
		panic(err)
	}

	orderService := order.New(log, storage, tokenTTL)

	grpcApp := grpcapp.New(log, orderService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
