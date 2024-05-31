package grpcapp

import (
	"fmt"
	orderGrpc "github.com/bxiit/order-service-pet-store/internal/grpc/order"
	orderv1 "github.com/bxiit/protos/gen/go/order"
	ssov1 "github.com/bxiit/protos/gen/go/sso"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

var AuthServiceClient ssov1.AuthClient
var OrderServiceClient orderv1.OrderServiceClient
var UserInfoServiceClient ssov1.UserInfoClient

func New(
	log *slog.Logger,
	catalogueService orderGrpc.OrderService,
	port int,
) *App {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			//logging.StartCall, logging.FinishCall,
			logging.PayloadReceived, logging.PayloadSent,
		),
		// Add any other option (check functions starting with logging.With).
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
		AdminInterceptorCreateItem,
		AdminInterceptorGetAllOrders,
		AdminInterceptorGetOrdersOfUser,
	))

	orderGrpc.Register(gRPCServer, catalogueService)

	return &App{
		log:        log,
		port:       port,
		grpcServer: gRPCServer,
	}
}

// MustRun runs gRPC server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run runs gRPC server.
func (a *App) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func ConnectToSsoService() {
	conn, err := grpc.NewClient("0.0.0.0:44044", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	AuthServiceClient = ssov1.NewAuthClient(conn)
	UserInfoServiceClient = ssov1.NewUserInfoClient(conn)
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.grpcServer.GracefulStop()
}
