package handler

import (
	"context"
	"encoding/json"
	"inventory-service/internal/cache"
	"inventory-service/internal/model"
	"inventory-service/internal/repository"
	pb "inventory-service/proto"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	Repo *repository.ProductRepository
	Cache *cache.RedisClient
}

func NewProductHandler(repo *repository.ProductRepository, cache *cache.RedisClient) *ProductHandler {
	return &ProductHandler{
		Repo:  repo,
		Cache: cache,
	}
}


func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.ProductInput) (*pb.Product, error) {
	id := primitive.NewObjectID().Hex()
	product := &model.Product{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
	}
	h.Repo.Create(ctx, product)
	return toProto(product), nil
}

func (h *ProductHandler) GetProductByID(ctx context.Context, req *pb.ProductID) (*pb.Product, error) {
	cacheKey := "product:" + req.Id

	// 1. Проверяем кеш
	cached, err := h.Cache.Get(cacheKey)
	if err == nil {
		var p pb.Product
		if err := json.Unmarshal([]byte(cached), &p); err == nil {
			log.Println("🔁 Отдано из Redis")
			return &p, nil
		}
	}

	// 2. Получаем из Mongo
	product, err := h.Repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	pbProduct := toProto(product)

	// 3. Кешируем
	jsonData, _ := json.Marshal(pbProduct)
	_ = h.Cache.Set(cacheKey, string(jsonData), time.Minute*5)

	log.Println("✅ Отдано из Mongo и закешировано")

	return pbProduct, nil
}


func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.Product) (*pb.Product, error) {
	p := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
	}
	h.Repo.Update(ctx, req.Id, p)
	p.ID = req.Id
	return toProto(p), nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.ProductID) (*pb.DeleteResponse, error) {
	err := h.Repo.Delete(ctx, req.Id)
	return &pb.DeleteResponse{Success: err == nil}, err
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *pb.ListRequest) (*pb.ProductList, error) {
	products, err := h.Repo.List(ctx, req.Category, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	var result []*pb.Product
	for _, p := range products {
		result = append(result, toProto(&p))
	}
	return &pb.ProductList{Products: result}, nil
}

func toProto(p *model.Product) *pb.Product {
	return &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Category:    p.Category,
	}
}
