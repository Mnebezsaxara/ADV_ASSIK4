package app

import (
	"log"
	"net"
	"order-service/internal/db"
	"order-service/internal/handler"
	"order-service/internal/repository"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	pb "order-service/proto"
)

func Run() {
	// Подключение к MongoDB
	client, err := db.NewMongoClient("mongodb://localhost:27017")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	db := client.Database("orderdb")
	repo := repository.NewOrderRepository(db)

	// Подключение к NATS
	natsConn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf(" Не удалось подключиться к NATS: %v", err)
	}
	log.Println(" Подключено к NATS")

	handler := handler.NewOrderHandler(repo, natsConn)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Failed to listen on port 50053: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, handler)

	log.Println(" Order gRPC сервер запущен на порту :50053")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
