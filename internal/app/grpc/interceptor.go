package grpcapp

import (
	"context"
	orderv1 "github.com/bxiit/protos/gen/go/order"
	ssov1 "github.com/bxiit/protos/gen/go/sso"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
)

type TokenClaims struct {
	UID   int64  `json:"uid"`
	Email string `json:"email"`
	AppID int    `json:"app_id"`
	jwt.MapClaims
}

func DecodeToken(appSecret string, tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func AdminInterceptorCreateItem(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/order.CatalogueService/CreateOrder" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}
	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}

	userInfo, err := UserInfoServiceClient.GetUserInfo(ctx, &ssov1.GetUserInfoRequest{Token: tkn[0]})
	if err != nil {
		log.Printf("failed to get user info from sso service")
		return nil, status.Errorf(codes.Internal, "failed to get user info from sso service")
	}

	if userInfo.User.Role != "admin" {
		return nil, status.Errorf(codes.Internal, "permission failed")
	}

	return handler(ctx, req)
}

func AdminInterceptorGetOrdersOfUser(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/order.OrderService/GetOrderByUserId" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}
	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}

	userInfo, err := UserInfoServiceClient.GetUserInfo(ctx, &ssov1.GetUserInfoRequest{Token: tkn[0]})
	if err != nil {
		log.Printf("failed to get user info from sso service")
		return nil, status.Errorf(codes.Internal, "failed to get user info from sso service")
	}

	var request orderv1.GetOrdersByUserId
	err = copier.Copy(&request, req)
	if err != nil {
		return nil, err
	}
	if userInfo.User.Id != request.UserId {
		return nil, status.Errorf(codes.InvalidArgument, "you can not get access to others orders")
	}

	isAdminRequest := &ssov1.IsAdminRequest{UserId: int64(userInfo.User.Id)}
	isAdminResponse, err := AuthServiceClient.IsAdmin(ctx, isAdminRequest)
	if err != nil {
		log.Printf("permissions fail %v", err)
		return nil, status.Errorf(codes.PermissionDenied, "permission failed")
	}

	if !isAdminResponse.IsAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "permission failed")
	}

	return handler(ctx, req)
}

func AdminInterceptorGetAllOrders(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/order.OrderService/ListOrders" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}
	tkn, found := md["authorization"]
	if !found || len(tkn) == 0 || tkn[0] == "" {
		return nil, status.Errorf(codes.PermissionDenied, "lack of permission")
	}

	userInfo, err := UserInfoServiceClient.GetUserInfo(ctx, &ssov1.GetUserInfoRequest{Token: tkn[0]})
	if err != nil {
		log.Printf("failed to get user info from sso service")
		return nil, status.Errorf(codes.Internal, "failed to get user info from sso service")
	}

	isAdminRequest := &ssov1.IsAdminRequest{UserId: int64(userInfo.User.Id)}
	isAdminResponse, err := AuthServiceClient.IsAdmin(ctx, isAdminRequest)
	if err != nil {
		log.Printf("permissions fail %v", err)
		return nil, status.Errorf(codes.Internal, "permission failed")
	}

	if !isAdminResponse.IsAdmin {
		return nil, status.Errorf(codes.Internal, "permission failed")
	}

	return handler(ctx, req)
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func OrderInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/order.OrderService/CreateOrder" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}

	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}

	isAuthenticatedRequest := &ssov1.IsAuthenticatedRequest{Token: tkn[0]}
	isAuthenticatedResponse, err := AuthServiceClient.IsAuthenticated(ctx, isAuthenticatedRequest)
	if err != nil {
		log.Printf("failed to get metadata from context")
		return nil, status.Errorf(codes.Internal, "failed to get metadata from context")
	}

	if !isAuthenticatedResponse.GetIsAuthenticated() {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}

	return handler(ctx, req)
}
