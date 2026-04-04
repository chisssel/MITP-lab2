package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UnixMilli(),
			"id":        uuid.New().String(),
			"data": gin.H{
				"items":  []string{"item1", "item2", "item3"},
				"count":  3,
				"nested": gin.H{"key": "value"},
			},
		})
	})

	r.POST("/echo", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, body)
	})

	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.Run(":8001")
}
