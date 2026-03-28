package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/ypapax/grpc_example/generated/order"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	resp, err := client.GetOrder(
		context.Background(),
		&pb.GetOrderRequest{
			OrderId: "order-123",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID:     %s\n", resp.Id)
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Total:  %.2f\n", resp.Total)
}
