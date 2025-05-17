package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"

	"user-service/internal/cache"
	"user-service/internal/model"
	"user-service/internal/repository"
	pb "user-service/proto"
)

type userServer struct {
	pb.UnimplementedUserServiceServer
	repo *repository.UserRepository
}

var rdb *cache.RedisClient

func (s *userServer) RegisterUser(ctx context.Context, req *pb.RegisterRequest) (*pb.UserResponse, error) {
	log.Printf("RegisterUser called: %s, %s", req.Username, req.Email)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	newID := primitive.NewObjectID().Hex()

	return &pb.UserResponse{
		Id:      newID,
		Message: "User successfully registered!",
	}, nil
}

func (s *userServer) AuthenticateUser(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return &pb.AuthResponse{Success: false}, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return &pb.AuthResponse{Success: false}, nil
	}

	return &pb.AuthResponse{
		Success: true,
		Token:   "mock-token-123",
	}, nil
}

func (s *userServer) GetUserProfile(ctx context.Context, req *pb.UserID) (*pb.UserProfile, error) {
	cacheKey := "user:" + req.Id

	cached, err := rdb.Get(cacheKey)
	if err == nil {
		var cachedUser pb.UserProfile
		if err := json.Unmarshal([]byte(cached), &cachedUser); err == nil {
			log.Println("🔁 Отдано из Redis кеша")
			return &cachedUser, nil
		}
	}

	user, err := s.repo.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	profile := &pb.UserProfile{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	jsonData, _ := json.Marshal(profile)
	_ = rdb.Set(cacheKey, string(jsonData), time.Minute*5)

	log.Println("✅ Отдано из Mongo и закешировано")

	return profile, nil
}

func main() {
	// Mongo
	client := repository.ConnectMongo()
	userRepo := repository.NewUserRepository(client)

	// Redis
	rdb = cache.NewRedisClient("localhost:6379")

	// gRPC
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, &userServer{
		repo: userRepo,
	})

	log.Println("✅ gRPC UserService запущен на :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
