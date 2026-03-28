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
	// Обязательный embed — без него не скомпилируется.
	// Выглядит как костыль: зачем встраивать заглушку
	// в продакшн-код? Но gRPC-Go это требует.
	//
	// Причина: gRPC родился внутри Google, где монорепо
	// и все сервисы обновляются одновременно. Если кто-то
	// добавит новый rpc метод в proto — сервер без этой
	// заглушки перестанет компилироваться. С заглушкой —
	// старые методы работают, новый вернёт ошибку
	// codes.Unimplemented пока не реализуешь.
	//
	// В обычных проектах (не монорепо) это менее актуально:
	// версия proto фиксирована в go.mod и сама не обновится.
	// Но таковы правила фреймворка.
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
