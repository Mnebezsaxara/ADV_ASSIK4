package handler

import (
	"context"
	// "io"
	"log"
	"net/http"
	"strconv"

	pb "api-gateway/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	pborder "api-gateway/proto/order"

)

const (
	inventoryServiceURL = "http://localhost:8081"
	orderServiceURL     = "http://localhost:8082"
)

var (
	userClient     pb.UserServiceClient
	inventoryClient pb.ProductServiceClient
	orderClient pborder.OrderServiceClient
)

func SetupRoutes(r *gin.Engine) {
	// Инициализация gRPC-клиента для UserService
	connUser, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf(" Не удалось подключиться к UserService: %v", err)
	}
	userClient = pb.NewUserServiceClient(connUser)
	log.Println("gRPC UserService клиент инициализирован")

	// Инициализация gRPC-клиента для InventoryService
	connInventory, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Не удалось подключиться к Inventory Service: %v", err)
	}
	inventoryClient = pb.NewProductServiceClient(connInventory)
	log.Println("gRPC Inventory клиент инициализирован")

	// Инициализация gRPC-клиента для InventoryService
	connOrder, err := grpc.Dial("localhost:50053", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Не удалось подключиться к OrderService: %v", err)
	}
	orderClient = pborder.NewOrderServiceClient(connOrder)
	log.Println("gRPC Order клиент инициализирован")


	// REST → gRPC маршруты для Users
	r.POST("/users/register", registerUser)
	r.POST("/users/login", loginUser)
	r.GET("/users/profile/:id", getUserProfile)

	// REST → gRPC маршруты для Products
	r.POST("/products", createProduct)
	r.GET("/products/:id", getProductByID)
	r.PATCH("/products/:id", updateProduct)
	r.DELETE("/products/:id", deleteProduct)
	r.GET("/products", listProducts)

	// r.Any("/categories/*any", proxy(inventoryServiceURL))
	// r.Any("/orders/*any", proxy(orderServiceURL))
	// r.Any("/orders", proxy(orderServiceURL))


	r.POST("/orders", createOrder)
	r.GET("/orders/:id", getOrderByID)
	r.PATCH("/orders/:id", updateOrderStatus)
	r.GET("/orders", getOrdersByUser)


	
}

// func proxy(target string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader != "Aldiyar2006" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 			return
// 		}
// 		url := target + c.Request.URL.Path
// 		if c.Request.URL.RawQuery != "" {
// 			url += "?" + c.Request.URL.RawQuery
// 		}
// 		req, err := http.NewRequest(c.Request.Method, url, c.Request.Body)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
// 			return
// 		}
// 		req.Header = c.Request.Header
// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		if err != nil {
// 			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach target service"})
// 			return
// 		}
// 		defer resp.Body.Close()
// 		body, _ := io.ReadAll(resp.Body)
// 		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
// 	}
// }

func registerUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := userClient.RegisterUser(context.Background(), &pb.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": resp.Id, "message": resp.Message})
}

func loginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	resp, err := userClient.AuthenticateUser(context.Background(), &pb.AuthRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil || !resp.Success {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "login successful", "token": resp.Token})
}

func getUserProfile(c *gin.Context) {
	userID := c.Param("id")
	resp, err := userClient.GetUserProfile(context.Background(), &pb.UserID{Id: userID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": resp.Id,
		"username": resp.Username,
		"email": resp.Email,
		"created_at": resp.CreatedAt,
	})
}

func createProduct(c *gin.Context) {
	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int32   `json:"stock"`
		Category    string  `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	res, err := inventoryClient.CreateProduct(context.Background(), &pb.ProductInput{
		Name: req.Name, Description: req.Description, Price: req.Price, Stock: req.Stock, Category: req.Category,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, res)
}

func getProductByID(c *gin.Context) {
	id := c.Param("id")
	res, err := inventoryClient.GetProductByID(context.Background(), &pb.ProductID{Id: id})
	if err != nil {
		c.JSON(404, gin.H{"error": "product not found"})
		return
	}
	c.JSON(200, res)
}

func updateProduct(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int32   `json:"stock"`
		Category    string  `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	res, err := inventoryClient.UpdateProduct(context.Background(), &pb.Product{
		Id: id, Name: req.Name, Description: req.Description, Price: req.Price, Stock: req.Stock, Category: req.Category,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, res)
}

func deleteProduct(c *gin.Context) {
	id := c.Param("id")
	res, err := inventoryClient.DeleteProduct(context.Background(), &pb.ProductID{Id: id})
	if err != nil || !res.Success {
		c.JSON(500, gin.H{"error": "failed to delete product"})
		return
	}
	c.JSON(200, gin.H{"status": "deleted"})
}

func listProducts(c *gin.Context) {
	category := c.Query("category")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, _ := strconv.ParseInt(pageStr, 10, 64)
	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	res, err := inventoryClient.ListProducts(context.Background(), &pb.ListRequest{
		Category: category, Page: page, Limit: limit,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list products"})
		return
	}
	c.JSON(200, res.Products)
}



func createOrder(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"`
		Products []struct {
			ProductID string `json:"product_id"`
			Quantity  int32  `json:"quantity"`
		} `json:"products"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	var grpcProducts []*pborder.OrderProductInput
	for _, p := range req.Products {
		grpcProducts = append(grpcProducts, &pborder.OrderProductInput{
			ProductId: p.ProductID,
			Quantity:  p.Quantity,
		})
	}

	res, err := orderClient.CreateOrder(context.Background(), &pborder.OrderInput{
		UserId:   req.UserID,
		Products: grpcProducts,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, res)
}

func getOrderByID(c *gin.Context) {
	id := c.Param("id")
	res, err := orderClient.GetOrderByID(context.Background(), &pborder.OrderID{Id: id})
	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}
	c.JSON(200, res)
}

func updateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	res, err := orderClient.UpdateOrderStatus(context.Background(), &pborder.OrderStatusUpdate{
		Id:     id,
		Status: req.Status,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, res)
}

func getOrdersByUser(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(400, gin.H{"error": "user_id is required"})
		return
	}
	res, err := orderClient.GetOrdersByUser(context.Background(), &pborder.UserID{UserId: userID})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, res.Orders)
}


// package handler

// import (
// 	"context"
// 	"io"
// 	"log"
// 	"net/http"

// 	pb "api-gateway/proto" // путь к сгенерированному .pb.go
// 	"github.com/gin-gonic/gin"
// 	"google.golang.org/grpc"
// )

// const (
// 	inventoryServiceURL = "http://localhost:8081"
// 	orderServiceURL     = "http://localhost:8082"
// )

// var userClient pb.UserServiceClient

// func SetupRoutes(r *gin.Engine) {
// 	// 🔌 Инициализация gRPC-клиента для UserService
// 	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("❌ Не удалось подключиться к UserService: %v", err)
// 	}
// 	userClient = pb.NewUserServiceClient(conn)
// 	log.Println("✅ gRPC UserService клиент инициализирован")

// 	// Inventory
// 	r.Any("/products/*any", proxy(inventoryServiceURL))
// 	r.Any("/categories/*any", proxy(inventoryServiceURL))
// 	r.Any("/products", proxy(inventoryServiceURL))

// 	// Orders
// 	r.Any("/orders/*any", proxy(orderServiceURL))
// 	r.Any("/orders", proxy(orderServiceURL))

// 	// Users
// 	r.POST("/users/register", registerUser)
// 	r.POST("/users/login", loginUser)
// 	r.GET("/users/profile/:id", getUserProfile)


// }

// func proxy(target string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Проверка авторизации
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader != "Aldiyar2006" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 			return
// 		}

// 		// Собираем полный URL для проксирования
// 		url := target + c.Request.URL.Path
// 		if c.Request.URL.RawQuery != "" {
// 			url += "?" + c.Request.URL.RawQuery
// 		}

// 		// Создаем новый HTTP-запрос
// 		req, err := http.NewRequest(c.Request.Method, url, c.Request.Body)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
// 			return
// 		}

// 		// Копируем заголовки
// 		req.Header = c.Request.Header

// 		// Отправляем запрос
// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		if err != nil {
// 			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach target service"})
// 			return
// 		}
// 		defer resp.Body.Close()

// 		// Читаем тело ответа
// 		body, _ := io.ReadAll(resp.Body)

// 		// Возвращаем ответ клиенту
// 		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
// 	}
// }

// // 🔐 Handler для POST /users/register
// func registerUser(c *gin.Context) {
// 	var req struct {
// 		Username string `json:"username"`
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
// 		return
// 	}

// 	resp, err := userClient.RegisterUser(context.Background(), &pb.RegisterRequest{
// 		Username: req.Username,
// 		Email:    req.Email,
// 		Password: req.Password,
// 	})

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"id":      resp.Id,
// 		"message": resp.Message,
// 	})
// }








// func loginUser(c *gin.Context) {
// 	var req struct {
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
// 		return
// 	}

// 	resp, err := userClient.AuthenticateUser(context.Background(), &pb.AuthRequest{
// 		Email:    req.Email,
// 		Password: req.Password,
// 	})

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if !resp.Success {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "login successful",
// 		"token":   resp.Token,
// 	})
// }




// func getUserProfile(c *gin.Context) {
// 	userID := c.Param("id")

// 	resp, err := userClient.GetUserProfile(context.Background(), &pb.UserID{
// 		Id: userID,
// 	})

// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"id":         resp.Id,
// 		"username":   resp.Username,
// 		"email":      resp.Email,
// 		"created_at": resp.CreatedAt,
// 	})
// }



















































// package handler

// import (
// 	"io"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
	
// )

// const (
// 	inventoryServiceURL = "http://localhost:8081" // порт Inventory-сервиса
// 	orderServiceURL     = "http://localhost:8082" // порт Order-сервиса
// )

// func SetupRoutes(r *gin.Engine) {
// 	// Inventory
// 	r.Any("/products/*any", proxy(inventoryServiceURL))
// 	r.Any("/categories/*any", proxy(inventoryServiceURL))





// 	r.Any("/products", proxy(inventoryServiceURL))  




// 	// Orders
// 	r.Any("/orders/*any", proxy(orderServiceURL))


// 	r.Any("/orders", proxy(orderServiceURL))
// 	// r.Any("/orders/:id", proxy(orderServiceURL))
// }

// func proxy(target string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Проверка авторизации
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader != "Aldiyar2006" {
// 			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
// 			return
// 		}

// 		// Собираем полный URL для проксирования
// 		url := target + c.Request.URL.Path
// 		if c.Request.URL.RawQuery != "" {
// 			url += "?" + c.Request.URL.RawQuery
// 		}

// 		// Создаем новый HTTP-запрос
// 		req, err := http.NewRequest(c.Request.Method, url, c.Request.Body)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
// 			return
// 		}

// 		// Копируем заголовки
// 		req.Header = c.Request.Header

// 		// Отправляем запрос
// 		client := &http.Client{}
// 		resp, err := client.Do(req)
// 		if err != nil {
// 			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach target service"})
// 			return
// 		}
// 		defer resp.Body.Close()

// 		// Читаем тело ответа
// 		body, _ := io.ReadAll(resp.Body)

// 		// Возвращаем ответ клиенту
// 		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
// 	}
// 
