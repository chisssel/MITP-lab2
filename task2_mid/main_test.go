package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(LoggerMiddleware())
	r.Use(gin.Recovery())
	r.GET("/hello", helloHandler)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	return r
}

func TestLoggerMiddleware(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHelloHandler(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantMsg    string
		wantErr    string
	}{
		{
			name:       "valid name John",
			query:      "?name=John",
			wantStatus: http.StatusOK,
			wantMsg:    "Hello, John!",
		},
		{
			name:       "valid name Alice",
			query:      "?name=Alice",
			wantStatus: http.StatusOK,
			wantMsg:    "Hello, Alice!",
		},
		{
			name:       "valid name with spaces",
			query:      "?name=John+Doe",
			wantStatus: http.StatusOK,
			wantMsg:    "Hello, John Doe!",
		},
		{
			name:       "name with digits",
			query:      "?name=Jo123hn",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Name cannot contain digits",
		},
		{
			name:       "name with only digits",
			query:      "?name=12345",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Name cannot contain digits",
		},
		{
			name:       "name missing",
			query:      "",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Name is required",
		},
		{
			name:       "empty name",
			query:      "?name=",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Name is required",
		},
		{
			name:       "whitespace only name",
			query:      "?name=   ",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Name is required",
		},
		{
			name:       "name with spaces and digits",
			query:      "?name=Jo 123",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Name cannot contain digits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/hello"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp HelloResponse
			json.Unmarshal(w.Body.Bytes(), &resp)

			if tt.wantMsg != "" && resp.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", resp.Message, tt.wantMsg)
			}

			if tt.wantErr != "" && resp.Error != tt.wantErr {
				t.Errorf("error = %q, want %q", resp.Error, tt.wantErr)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantErr string
	}{
		{"valid name", "John", true, ""},
		{"valid name with space", "John Doe", true, ""},
		{"name with digits", "Jo123hn", false, "Name cannot contain digits"},
		{"empty string", "", false, "Name is required"},
		{"whitespace only", "   ", false, "Name is required"},
		{"name with leading spaces", "  John", true, ""},
		{"name with trailing spaces", "John  ", true, ""},
		{"name starting with digit", "1John", false, "Name cannot contain digits"},
		{"name ending with digit", "John1", false, "Name cannot contain digits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, errMsg := validateName(tt.input)
			if ok != tt.wantOk {
				t.Errorf("validateName(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if !ok && errMsg != tt.wantErr {
				t.Errorf("validateName(%q) err = %q, want %q", tt.input, errMsg, tt.wantErr)
			}
		})
	}
}

func TestPingEndpoint(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "pong" {
		t.Errorf("message = %q, want %q", resp["message"], "pong")
	}
}

func TestContentType(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"/ping", "/ping", "application/json"},
		{"/hello valid", "/hello?name=John", "application/json"},
		{"/hello invalid", "/hello?name=123", "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, tt.want) {
				t.Errorf("Content-Type = %q, want %q", contentType, tt.want)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
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

	req, _ := http.NewRequest("POST", "/hello?name=John", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("POST to /hello should not return 200")
	}
}

func TestSpecialCharactersInName(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{"underscore in name", "?name=John_Doe", http.StatusOK},
		{"hyphen in name", "?name=John-Doe", http.StatusOK},
		{"dot in name", "?name=John.Doe", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/hello"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestLogFormat(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)

	router := setupRouter()
	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logOutput := logBuf.String()

	logPattern := regexp.MustCompile(`\[GET\] /ping \| Status: 200 \| Latency: .+ \| IP:`)
	if !logPattern.MatchString(logOutput) {
		t.Errorf("log output = %q, does not match expected format", logOutput)
	}
}

func TestLogRequest(t *testing.T) {
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(nil)

	logRequest("GET", "/test", 200, 100*time.Millisecond, "127.0.0.1")

	logOutput := logBuf.String()

	expectedParts := []string{"GET", "/test", "200", "127.0.0.1"}
	for _, part := range expectedParts {
		if !strings.Contains(logOutput, part) {
			t.Errorf("log missing part: %s, full log: %s", part, logOutput)
		}
	}
}

func TestLogFile(t *testing.T) {
	origFileLogger := fileLogger
	origLogFile := logFile
	defer func() {
		fileLogger = origFileLogger
		logFile = origLogFile
	}()

	testLogFile := "test_server.log"
	defer os.Remove(testLogFile)

	if err := initLogger(testLogFile); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	defer closeLogger()

	logRequest("GET", "/ping", 200, 10*time.Millisecond, "127.0.0.1")
	fileLogger.Println("test entry")

	content, err := os.ReadFile(testLogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "GET") {
		t.Errorf("Log file should contain GET, got: %s", string(content))
	}
}
