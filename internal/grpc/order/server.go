package orderGrpc

import (
	"context"
	"github.com/bxiit/order-service-pet-store/internal/data/dto"
	"github.com/bxiit/order-service-pet-store/internal/data/models"
	"github.com/bxiit/protos/gen/go/catalogue"
	orderv20 "github.com/bxiit/protos/gen/go/order"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

type OrderService interface {
	CreateOrder(context.Context, *dto.OrderDTO) (*dto.OrderDTO, error)
	ListOrders(context.Context) ([]*models.Order, error)
	GetOrder(context.Context, int) (*models.Order, error)
	GetOrdersByUserId(context.Context, int) ([]*dto.OrderDTO, error)
}

type orderService struct {
	orderv20.UnimplementedOrderServiceServer
	order OrderService
}

func Register(gRPCServer *grpc.Server, order OrderService) {
	orderv20.RegisterOrderServiceServer(gRPCServer, &orderService{order: order})
}

func (os *orderService) CreateOrder(ctx context.Context, req *orderv20.CreateOrderRequest) (*orderv20.CreateOrderResponse, error) {
	var ord dto.OrderDTO
	err := copier.Copy(&ord, req.Order)
	if err != nil {
		log.Fatalf("failed to copy %v", err)
		return nil, err
	}

	orderDTO, err := os.order.CreateOrder(ctx, &ord)
	if err != nil {
		return nil, status.Error(codes.Internal, "error with create ord")
	}

	var response orderv20.Order
	err = copier.Copy(&response, orderDTO)
	if err != nil {
		return nil, status.Error(codes.Internal, "error with copying to dto")
	}

	return &orderv20.CreateOrderResponse{Order: &response}, nil
}

func (os *orderService) ListOrders(ctx context.Context, req *orderv20.ListOrdersRequest) (*orderv20.ListOrdersResponse, error) {
	var responseOrders []*orderv20.Order

	orders, err := os.order.ListOrders(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range orders {
		var ord orderv20.Order
		err := copier.Copy(&ord, &item)
		if err != nil {
			log.Fatalf("failed to copy %v", err)
			return nil, err
		}

		responseOrders = append(responseOrders, &ord)
	}

	return &orderv20.ListOrdersResponse{Orders: responseOrders}, nil
}

func (os *orderService) GetOrder(ctx context.Context, req *orderv20.GetOrderRequest) (*orderv20.GetOrderResponse, error) {
	id, err := strconv.Atoi(req.Id)
	if err != nil {
		return nil, err
	}

	order, err := os.order.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	orderResponse := &orderv20.Order{
		Id:     order.ID,
		ItemId: order.ItemId,
		UserId: order.UserId,
	}
	return &orderv20.GetOrderResponse{Order: orderResponse}, nil
}

func (os *orderService) GetOrderByUserId(ctx context.Context, req *orderv20.GetOrdersByUserId) (*orderv20.ListOrdersResponse, error) {
	userId := req.GetUserId()
	if userId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	ordersByUserId, err := os.order.GetOrdersByUserId(ctx, int(userId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get orders of user")
	}

	var ordersResponse []*orderv20.Order
	for _, order := range ordersByUserId {
		var o orderv20.Order
		err := copier.Copy(&o, order)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to copy to response order")
		}

		ordersResponse = append(ordersResponse, &o)
	}

	var items []*cataloguev20.Item
	items = append(items, &cataloguev20.Item{Id: 1})
	items = append(items, &cataloguev20.Item{Id: 2})
	items = append(items, &cataloguev20.Item{Id: 3})
	items = append(items, &cataloguev20.Item{Id: 4})

	return &orderv20.ListOrdersResponse{Orders: ordersResponse}, nil
}
