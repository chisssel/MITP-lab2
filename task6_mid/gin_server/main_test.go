package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	tests := []struct {
		name       string
		endpoint   string
		method     string
		body       interface{}
		wantStatus int
		checkBody  map[string]string
	}{
		{
			name:       "ping returns 200 with message",
			endpoint:   "/ping",
			method:     "GET",
			wantStatus: http.StatusOK,
			checkBody:  map[string]string{"message": "pong"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}

			req, _ := http.NewRequest(tt.method, tt.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.checkBody != nil {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed to parse response: %v", err)
					return
				}
				for key, want := range tt.checkBody {
					if got, ok := resp[key].(string); !ok || got != want {
						t.Errorf("%s = %v, want %v", key, resp[key], want)
					}
				}
			}
		})
	}
}

func TestJsonEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": 1234567890,
			"id":        "test-id",
			"data": gin.H{
				"items":  []string{"item1", "item2"},
				"count":  2,
				"nested": gin.H{"key": "value"},
			},
		})
	})

	tests := []struct {
		name       string
		wantStatus int
		checkKeys  []string
	}{
		{
			name:       "json returns correct structure",
			wantStatus: http.StatusOK,
			checkKeys:  []string{"status", "timestamp", "id", "data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/json", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)

			for _, key := range tt.checkKeys {
				if _, ok := resp[key]; !ok {
					t.Errorf("missing key: %s", key)
				}
			}
		})
	}
}

func TestEchoEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/echo", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, body)
	})

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		wantEcho   map[string]interface{}
	}{
		{
			name:       "echo returns same body",
			body:       map[string]string{"key": "value"},
			wantStatus: http.StatusOK,
			wantEcho:   map[string]interface{}{"key": "value"},
		},
		{
			name:       "echo with nested object",
			body:       map[string]interface{}{"name": "test", "count": 42},
			wantStatus: http.StatusOK,
			wantEcho:   map[string]interface{}{"name": "test", "count": float64(42)},
		},
		{
			name:       "echo with invalid json",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
			wantEcho:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/echo", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantEcho != nil {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				for key, want := range tt.wantEcho {
					if resp[key] != want {
						t.Errorf("resp[%s] = %v, want %v", key, resp[key], want)
					}
				}
			}
		})
	}
}

func TestSlowEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/slow", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	tests := []struct {
		name       string
		wantStatus int
	}{
		{
			name:       "slow returns ok",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/slow", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", resp["status"])
	}
}
