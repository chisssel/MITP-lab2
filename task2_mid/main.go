package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	fileLogger *log.Logger
	logFile    *os.File
)

func initLogger(filename string) error {
	var err error
	logFile, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	fileLogger = log.New(logFile, "", 0)
	return nil
}

func closeLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func logRequest(method, path string, status int, latency time.Duration, ip string) {
	logEntry := fmt.Sprintf("[%s] %s | Status: %d | Latency: %v | IP: %s",
		method, path, status, latency, ip)

	log.Println(logEntry)
	if fileLogger != nil {
		fileLogger.Println(logEntry)
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		logRequest(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
		)
	}
}

type HelloResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

var digitsRegex = regexp.MustCompile(`[0-9]`)

func validateName(name string) (bool, string) {
	name = strings.TrimSpace(name)

	if name == "" {
		return false, "Name is required"
	}

	if name == "" {
		return false, "Name is empty"
	}

	if digitsRegex.MatchString(name) {
		return false, "Name cannot contain digits"
	}

	return true, ""
}

func helloHandler(c *gin.Context) {
	name := c.Query("name")

	valid, errMsg := validateName(name)
	if !valid {
		c.JSON(http.StatusBadRequest, HelloResponse{Error: errMsg})
		return
	}

	c.JSON(http.StatusOK, HelloResponse{Message: "Hello, " + name + "!"})
}

func main() {
	if err := initLogger("server.log"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer closeLogger()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(LoggerMiddleware())
	r.Use(gin.Recovery())

	r.GET("/hello", helloHandler)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	log.Println("Server starting on :8080")
	log.Println("Logs are saved to: server.log")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
