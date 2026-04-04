package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DataNested struct {
	Key string `json:"key"`
}

type DataItems struct {
	Items  []string   `json:"items"`
	Count  int        `json:"count"`
	Nested DataNested `json:"nested"`
}

type JsonResponse struct {
	Status    string    `json:"status"`
	Timestamp int64     `json:"timestamp"`
	ID        string    `json:"id"`
	Data      DataItems `json:"data"`
}

type SlowResponse struct {
	Status string `json:"status"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

var OpenAPISpec = map[string]interface{}{
	"openapi": "3.0.0",
	"info": map[string]interface{}{
		"title":       "Gin Server API",
		"description": "REST API with OpenAPI documentation",
		"version":     "1.0.0",
	},
	"servers": []map[string]string{
		{"url": "http://localhost:8001", "description": "Gin server"},
	},
	"paths": map[string]interface{}{
		"/ping": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Ping endpoint",
				"description": "Returns a simple pong message for health checks",
				"tags":        []string{"Health"},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful response",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"message": map[string]string{"type": "string", "example": "pong"},
									},
								},
							},
						},
					},
				},
			},
		},
		"/json": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "JSON endpoint",
				"description": "Returns a complex JSON response with UUID and nested data",
				"tags":        []string{"Data"},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful response",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/JsonResponse",
								},
							},
						},
					},
				},
			},
		},
		"/echo": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "Echo endpoint",
				"description": "Returns the same JSON body that was sent in the request",
				"tags":        []string{"Data"},
				"requestBody": map[string]interface{}{
					"required":    true,
					"description": "JSON body to echo",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Echoed body"},
					"400": map[string]interface{}{"description": "Invalid JSON body"},
				},
			},
		},
		"/slow": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Slow endpoint",
				"description": "Returns a response after a 50ms delay",
				"tags":        []string{"Data"},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful response",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/SlowResponse",
								},
							},
						},
					},
				},
			},
		},
		"/health": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Health check",
				"description": "Returns the health status of the server",
				"tags":        []string{"Health"},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Server is healthy",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/HealthResponse",
								},
							},
						},
					},
				},
			},
		},
	},
	"components": map[string]interface{}{
		"schemas": map[string]interface{}{
			"JsonResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status":    map[string]string{"type": "string", "example": "ok"},
					"timestamp": map[string]string{"type": "integer", "example": "1234567890"},
					"id":        map[string]string{"type": "string", "example": "550e8400-e29b-41d4-a716-446655440000"},
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"items": map[string]interface{}{
								"type":  "array",
								"items": map[string]string{"type": "string"},
							},
							"count": map[string]string{"type": "integer"},
							"nested": map[string]interface{}{
								"$ref": "#/components/schemas/Nested",
							},
						},
					},
				},
			},
			"Nested": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]string{"type": "string", "example": "value"},
				},
			},
			"SlowResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]string{"type": "string", "example": "ok"},
				},
			},
			"HealthResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]string{"type": "string", "example": "healthy"},
				},
			},
		},
	},
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, JsonResponse{
			Status:    "ok",
			Timestamp: 0,
			ID:        uuid.New().String(),
			Data: DataItems{
				Items:  []string{"item1", "item2", "item3"},
				Count:  3,
				Nested: DataNested{Key: "value"},
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
		c.JSON(http.StatusOK, SlowResponse{Status: "ok"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{Status: "healthy"})
	})

	r.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, OpenAPISpec)
	})

	r.Run(":8001")
}
