package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

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

	return r
}

func TestPingEndpoint(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		checkBody  map[string]string
	}{
		{"GET /ping returns 200", "GET", "/ping", http.StatusOK, map[string]string{"message": "pong"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			for key, want := range tt.checkBody {
				if got, ok := resp[key].(string); !ok || got != want {
					t.Errorf("%s = %v, want %v", key, resp[key], want)
				}
			}
		})
	}
}

func TestJsonEndpoint(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		wantStatus int
		checkKeys  []string
	}{
		{"JSON returns correct keys", http.StatusOK, []string{"status", "timestamp", "id", "data"}},
		{"JSON data has required fields", http.StatusOK, []string{"items", "count", "nested"}},
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
	router := setupRouter()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		wantKeys   []string
	}{
		{"echo simple object", map[string]string{"key": "value"}, http.StatusOK, []string{"key"}},
		{"echo nested object", map[string]interface{}{"name": "test", "count": 42}, http.StatusOK, []string{"name", "count"}},
		{"echo empty object", map[string]string{}, http.StatusOK, []string{}},
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

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			for _, key := range tt.wantKeys {
				if _, ok := resp[key]; !ok {
					t.Errorf("missing key in response: %s", key)
				}
			}
		})
	}
}

func TestEchoInvalidJSON(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("POST", "/echo", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestSlowEndpoint(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp SlowResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Status != "ok" {
		t.Errorf("status = %v, want ok", resp.Status)
	}
}

func TestHealthEndpoint(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   map[string]string
	}{
		{"health returns healthy", "GET", "/health", http.StatusOK, map[string]string{"status": "healthy"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			for key, want := range tt.wantBody {
				if got, ok := resp[key].(string); !ok || got != want {
					t.Errorf("%s = %v, want %v", key, resp[key], want)
				}
			}
		})
	}
}

func TestOpenAPIEndpoint(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantKeys   []string
	}{
		{"OpenAPI returns 200", "/openapi.json", http.StatusOK, []string{"openapi", "info", "paths"}},
		{"OpenAPI has version", "/openapi.json", http.StatusOK, []string{"openapi"}},
		{"OpenAPI has paths", "/openapi.json", http.StatusOK, []string{"paths"}},
		{"OpenAPI has components", "/openapi.json", http.StatusOK, []string{"components"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %s, want application/json", contentType)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			for _, key := range tt.wantKeys {
				if _, ok := resp[key]; !ok {
					t.Errorf("missing key in OpenAPI spec: %s", key)
				}
			}
		})
	}
}

func TestOpenAPIPathsDefinition(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/openapi.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var spec map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &spec)

	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("paths is not a map")
	}

	expectedPaths := []string{"/ping", "/json", "/echo", "/slow", "/health"}
	for _, path := range expectedPaths {
		if _, exists := paths[path]; !exists {
			t.Errorf("missing path in OpenAPI: %s", path)
		}
	}
}

func TestOpenAPISchemas(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/openapi.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var spec map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &spec)

	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatal("components not found")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		t.Fatal("schemas not found")
	}

	expectedSchemas := []string{"JsonResponse", "SlowResponse", "HealthResponse", "Nested"}
	for _, schema := range expectedSchemas {
		if _, exists := schemas[schema]; !exists {
			t.Errorf("missing schema in OpenAPI: %s", schema)
		}
	}
}

func TestNotFoundEndpoint(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("POST", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}
