	package main

	import (
		"api-gateway/internal/handler"
		"github.com/gin-gonic/gin"
	)

	func main() {
		r := gin.Default()

		handler.SetupRoutes(r)

		r.Run(":8080") 
	}
