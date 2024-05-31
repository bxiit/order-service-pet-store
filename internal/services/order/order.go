package order

import (
	"context"
	"encoding/json"
	"fmt"
	grpcapp "github.com/bxiit/order-service-pet-store/internal/app/grpc"
	"github.com/bxiit/order-service-pet-store/internal/data/dto"
	"github.com/bxiit/order-service-pet-store/internal/data/models"
	"github.com/bxiit/order-service-pet-store/internal/sl"
	ssov1 "github.com/bxiit/protos/gen/go/sso"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"time"
)

type Order struct {
	log           *slog.Logger
	orderProvider OrderRepo
	channel       *amqp.Channel
	tokenTTL      time.Duration
}

func New(
	log *slog.Logger,
	orderProvider OrderRepo,
	tokenTtl time.Duration,
) *Order {
	return &Order{
		log:           log,
		orderProvider: orderProvider,
		tokenTTL:      tokenTtl,
	}
}

type OrderRepo interface {
	SaveOrder(ctx context.Context, orderDTO *dto.OrderDTO) (*dto.OrderDTO, error)
	GetAllOrders(context.Context) ([]*models.Order, error)
	GetOrderById(context.Context, int) (*models.Order, error)
	GetOrdersByUserId(context.Context, int) ([]*dto.OrderDTO, error)
}

func (o *Order) CreateOrder(ctx context.Context, orderDTO *dto.OrderDTO) (*dto.OrderDTO, error) {
	const op = "Order.CreateOrder"

	log := o.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to create orderDTO")

	orderDTO, err := o.orderProvider.SaveOrder(ctx, orderDTO)
	if err != nil {
		o.log.Warn("failed to save orderDTO", sl.Err(err))
		return nil, fmt.Errorf("%s", op)
	}

	err = sendNotification(ctx, orderDTO)
	if err != nil {
		o.log.Warn("failed to publish message", sl.Err(err))
		return nil, fmt.Errorf("%s", op)
	}

	return orderDTO, nil
}

func sendNotification(ctx context.Context, orderDTO *dto.OrderDTO) error {
	const op = "Order.sendNotification"
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("%s", op)
	}

	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return status.Errorf(codes.Unauthenticated, "authentication is required")
	}
	userInfo, err := grpcapp.UserInfoServiceClient.GetUserInfo(ctx, &ssov1.GetUserInfoRequest{Token: tkn[0]})
	if err != nil {
		log.Printf("failed to get user info from sso service")
		return status.Errorf(codes.Internal, "failed to get user info from sso service")
	}
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Print("failed to connect in new connection", sl.Err(err))
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Print("failed to create channel in new MQ", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	q, err := ch.QueueDeclare(
		"order", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Print("failed to declare queue", sl.Err(err))
		return fmt.Errorf("%s", op)
	}

	data := map[string]interface{}{
		"user_info":  userInfo.User,
		"order_info": orderDTO,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Print("Failed to marshal data", sl.Err(err))
		return fmt.Errorf("%s", op)
	}
	body := dataBytes
	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	return nil
}

func (o *Order) ListOrders(ctx context.Context) ([]*models.Order, error) {
	const op = "Order.ListOrders"
	log := o.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to get all orders")

	items, err := o.orderProvider.GetAllOrders(ctx)
	if err != nil {
		o.log.Warn("failed to get all items", sl.Err(err))
		return nil, err
	}

	return items, nil
}

func (o *Order) GetOrder(ctx context.Context, id int) (*models.Order, error) {
	const op = "Order.GetOrder"
	log := o.log.With(
		slog.String("op", op),
		slog.Int("item id", id),
	)

	log.Info("attempting to get order")
	item, err := o.orderProvider.GetOrderById(ctx, id)
	if err != nil {
		o.log.Warn("failed to get order", sl.Err(err))
		return nil, err
	}

	return item, nil
}

func (o *Order) GetOrdersByUserId(ctx context.Context, userId int) ([]*dto.OrderDTO, error) {
	const op = "Order.GetOrder"
	log := o.log.With(
		slog.String("op", op),
		slog.Int("user id", userId),
	)

	log.Info("attempting to get orders of user with id ", userId)
	ordersByUserId, err := o.orderProvider.GetOrdersByUserId(ctx, userId)
	if err != nil {
		o.log.Warn("failed to get orders of user", sl.Err(err))
		return nil, err
	}

	return ordersByUserId, nil
}
