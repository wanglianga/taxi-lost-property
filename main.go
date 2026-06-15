package main

import (
	"log"
	"os"
	"taxi-lost-property/internal/handler"
	"taxi-lost-property/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := repository.InitDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	log.Println("数据库初始化成功")

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		reports := api.Group("/reports")
		{
			reports.POST("", handler.CreateLostReport)
			reports.GET("", handler.ListLostReports)
			reports.GET("/:id", handler.GetLostReport)
			reports.POST("/:id/match", handler.MatchOrders)
			reports.POST("/:id/confirm-order", handler.ConfirmMatchedOrder)
		}

		submissions := api.Group("/submissions")
		{
			submissions.POST("", handler.CreateDriverSubmission)
			submissions.GET("", handler.ListDriverSubmissions)
			submissions.GET("/:id", handler.GetDriverSubmission)
		}

		inventory := api.Group("/inventory")
		{
			inventory.POST("", handler.CreateInventory)
			inventory.GET("", handler.ListInventory)
			inventory.GET("/:id", handler.GetInventory)
		}

		stations := api.Group("/stations")
		{
			stations.GET("", handler.ListStations)
			stations.GET("/:id", handler.GetStation)
		}

		claims := api.Group("/claims")
		{
			claims.POST("", handler.CreateClaim)
			claims.GET("", handler.ListClaims)
			claims.GET("/:id", handler.GetClaim)
			claims.POST("/:id/verify", handler.VerifyClaim)
			claims.POST("/:id/return", handler.ConfirmReturn)
		}

		orders := api.Group("/orders")
		{
			orders.GET("", handler.ListOrders)
			orders.GET("/:id", handler.GetOrder)
		}

		fuzzy := api.Group("/fuzzy")
		{
			fuzzy.POST("/search-vehicles", handler.FuzzySearchVehicles)
		}

		verify := api.Group("/verify")
		{
			verify.GET("/requirements", handler.GetVerifyRequirements)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "taxi-lost-property"})
	})

	port := ":8080"
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	}
	log.Printf("服务启动于 %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
