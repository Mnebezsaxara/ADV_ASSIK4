package app

import (
	"encoding/json"
	"inventory-service/internal/cache"
	"inventory-service/internal/db"
	"inventory-service/internal/handler"
	"inventory-service/internal/repository"
	"log"
	"net"

	pb "inventory-service/proto"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func Run() {
	// Подключение к MongoDB
	client, err := db.NewMongoClient("mongodb://192.168.1.70:27017/?replicaSet=rs0")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	db := client.Database("inventory")
	repo := repository.NewProductRepository(db)

	// ✅ Инициализация Redis
	redisClient := cache.NewRedisClient("localhost:6379")

	// gRPC handler с Redis
	handler := handler.NewProductHandler(repo, redisClient)

	// Подключение к NATS
	natsConn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf(" NATS: %v", err)
	}
	log.Println(" Подключено к NATS (Inventory)")

	// Подписка на события
	natsConn.Subscribe("order.created", func(m *nats.Msg) {
		log.Println(" Получено событие order.created")

		var order struct {
			OrderID string `json:"order_id"`
			UserID  string `json:"user_id"`
			Products []struct {
				ProductID string `json:"product_id"`
				Quantity  int32  `json:"quantity"`
			} `json:"products"`
		}

		if err := json.Unmarshal(m.Data, &order); err != nil {
			log.Println(" Ошибка парсинга события:", err)
			return
		}

		for _, p := range order.Products {
			err := repo.DecreaseStock(p.ProductID, p.Quantity)
			if err != nil {
				log.Printf(" Не удалось обновить stock %s: %v", p.ProductID, err)
			} else {
				log.Printf(" Stock is reduced: %s : -%d", p.ProductID, p.Quantity)
			}
		}
	})

	// Запуск gRPC-сервера
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen on port 50052: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, handler)

	log.Println(" Inventory gRPC сервер запущен на порту :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
