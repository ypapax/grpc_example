package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/ypapax/grpc_example/generated/order"

	"google.golang.org/grpc"
)

type server struct {
	// Обязательный embed.
	// Без него не скомпилируется.
	// Выглядит как костыль:
	// заглушка в продакшн-коде.
	// Но gRPC-Go это требует.
	//
	// gRPC родился в Google
	// (монорепо, всё обновляется
	// разом). Добавили новый rpc —
	// старые серверы не ломаются,
	// новый метод вернёт ошибку
	// codes.Unimplemented.
	//
	// В обычных проектах менее
	// актуально: версия proto
	// фиксирована в go.mod.
	// Но правила фреймворка.
	pb.UnimplementedOrderServiceServer
}

func (s *server) GetOrder(
	ctx context.Context,
	req *pb.GetOrderRequest,
) (*pb.GetOrderResponse, error) {
	fmt.Printf("Получил запрос: order_id=%s\n", req.OrderId)
	return &pb.GetOrderResponse{
		Id:     req.OrderId,
		Status: "paid",
		Total:  1500.0,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Сервер запущен на :50051")

	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
