package handler

import (
	"context"
	"encoding/json"
	"log"
	"order-service/internal/model"
	"order-service/internal/repository"
	pb "order-service/proto"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	Repo     *repository.OrderRepository
	NatsConn *nats.Conn
}

func NewOrderHandler(repo *repository.OrderRepository, natsConn *nats.Conn) *OrderHandler {
	return &OrderHandler{Repo: repo, NatsConn: natsConn}
}
func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.OrderInput) (*pb.Order, error) {
	products := make([]model.OrderProduct, len(req.Products))
	for i, p := range req.Products {
		products[i] = model.OrderProduct{ProductID: p.ProductId, Quantity: p.Quantity}
	}
	order := &model.Order{
		UserID:   req.UserId,
		Products: products,
	}

	// 💥 Транзакция
	client := h.Repo.Collection.Database().Client()
	session, err := client.StartSession()
	if err != nil {
		log.Println("❌ Ошибка запуска транзакции:", err)
		return nil, err
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		if err := h.Repo.CreateWithTx(sc, session, order); err != nil {
			_ = session.AbortTransaction(sc)
			return err
		}

		// 📨 Публикация события
		eventProducts := make([]map[string]interface{}, len(products))
		for i, p := range products {
			eventProducts[i] = map[string]interface{}{
				"product_id": p.ProductID,
				"quantity":   p.Quantity,
			}
		}
		eventData := map[string]interface{}{
			"order_id": order.ID.Hex(),
			"user_id":  order.UserID,
			"products": eventProducts,
		}
		payload, _ := json.Marshal(eventData)
		err = h.NatsConn.Publish("order.created", payload)
		if err != nil {
			_ = session.AbortTransaction(sc)
			log.Println("❌ Ошибка публикации события в NATS:", err)
			return err
		}

		log.Println("📤 Событие 'order.created' отправлено в NATS")

		if err := session.CommitTransaction(sc); err != nil {
			return err
		}

		log.Println("✅ Транзакция успешно завершена")
		return nil
	})

	if err != nil {
		log.Println("❌ Ошибка транзакции:", err)
		return nil, err
	}

	return toProto(order), nil
}


func (h *OrderHandler) GetOrderByID(ctx context.Context, req *pb.OrderID) (*pb.Order, error) {
	order, err := h.Repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toProto(order), nil
}

func (h *OrderHandler) UpdateOrderStatus(ctx context.Context, req *pb.OrderStatusUpdate) (*pb.Order, error) {
	h.Repo.UpdateStatus(ctx, req.Id, req.Status)
	order, _ := h.Repo.GetByID(ctx, req.Id)
	return toProto(order), nil
}

func (h *OrderHandler) GetOrdersByUser(ctx context.Context, req *pb.UserID) (*pb.OrderList, error) {
	orders, err := h.Repo.GetByUserID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	var protoOrders []*pb.Order
	for _, o := range orders {
		protoOrders = append(protoOrders, toProto(&o))
	}
	return &pb.OrderList{Orders: protoOrders}, nil
}

func toProto(o *model.Order) *pb.Order {
	products := make([]*pb.OrderProduct, len(o.Products))
	for i, p := range o.Products {
		products[i] = &pb.OrderProduct{ProductId: p.ProductID, Quantity: p.Quantity}
	}
	return &pb.Order{
		Id:        o.ID.Hex(),
		UserId:    o.UserID,
		Products:  products,
		Status:    o.Status,
		CreatedAt: o.CreatedAt,
	}
}