package repository

import (
	"context"
	"inventory-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductRepository struct {
	Collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) *ProductRepository {
	return &ProductRepository{Collection: db.Collection("products")}
}

func (r *ProductRepository) Create(ctx context.Context, p *model.Product) error {
	_, err := r.Collection.InsertOne(ctx, p)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	var product model.Product
	err := r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	return &product, err
}

func (r *ProductRepository) Update(ctx context.Context, id string, p *model.Product) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err := r.Collection.UpdateByID(ctx, objID, bson.M{"$set": p})
	return err
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err := r.Collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *ProductRepository) List(ctx context.Context, category string, page int64, limit int64) ([]model.Product, error) {
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}

	skip := (page - 1) * limit
	opts := options.Find().SetLimit(limit).SetSkip(skip)

	cursor, err := r.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var products []model.Product
	err = cursor.All(ctx, &products)
	return products, err
}

func (r *ProductRepository) DecreaseStock(productID string, quantity int32) error {
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$inc": bson.M{"stock": -quantity}}
	_, err = r.Collection.UpdateOne(context.TODO(), filter, update)
	return err
}