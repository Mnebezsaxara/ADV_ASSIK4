package repository

import (
	"context"
	"order-service/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository struct {
	Collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{Collection: db.Collection("orders")}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now().Unix()
	order.Status = "pending"
	_, err := r.Collection.InsertOne(ctx, order)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	var order model.Order
	err := r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&order)
	return &order, err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err := r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": status}})
	return err
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID string) ([]model.Order, error) {
	cursor, err := r.Collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	var orders []model.Order
	err = cursor.All(ctx, &orders)
	return orders, err
}

func (r *OrderRepository) CreateWithTx(ctx context.Context, session mongo.Session, order *model.Order) error {
	order.ID = primitive.NewObjectID()
	order.CreatedAt = time.Now().Unix()
	order.Status = "pending"

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := r.Collection.FindOne(sc, bson.M{"_id": order.ID}).Err(); err == nil {
			return mongo.ErrClientDisconnected // пример: защита от дублирования
		}
		_, err := r.Collection.InsertOne(sc, order)
		return err
	})
}
