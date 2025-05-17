// package model

// import (
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type Order struct {
// 	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
// 	UserID    string             `json:"user_id" bson:"user_id"`
// 	Products  []OrderProduct     `json:"products" bson:"products"`
// 	Status    string             `json:"status" bson:"status"` // pending, completed, cancelled
// 	CreatedAt int64              `json:"created_at" bson:"created_at"`
// }

// type OrderProduct struct {
// 	ProductID string `json:"product_id" bson:"product_id"`
// 	Quantity  int    `json:"quantity" bson:"quantity"`
// }

// package model

// type Order struct {
// 	ID        string        `bson:"_id,omitempty"`
// 	UserID    string        `bson:"user_id"`
// 	Products  []OrderProduct `bson:"products"`
// 	Status    string        `bson:"status"`
// 	CreatedAt int64         `bson:"created_at"`
// }

// type OrderProduct struct {
// 	ProductID string `bson:"product_id"`
// 	Quantity  int32  `bson:"quantity"`
// }
package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Order struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    string             `bson:"user_id"`
	Products  []OrderProduct     `bson:"products"`
	Status    string             `bson:"status"`
	CreatedAt int64              `bson:"created_at"`
}

type OrderProduct struct {
	ProductID string `bson:"product_id"`
	Quantity  int32  `bson:"quantity"`
}