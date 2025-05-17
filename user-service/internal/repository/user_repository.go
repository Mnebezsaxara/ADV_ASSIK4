package repository

import (
	"context"
	"user-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
    Collection *mongo.Collection
}

func NewUserRepository(client *mongo.Client) *UserRepository {
    return &UserRepository{
        Collection: client.Database("userdb").Collection("users"),
    }
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
    _, err := r.Collection.InsertOne(ctx, user)
    return err
}





func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
    var user model.User
    err := r.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}





func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }

    var user model.User
    err = r.Collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
    if err != nil {
        return nil, err
    }

    return &user, nil
}
