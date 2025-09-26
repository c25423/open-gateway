package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/c25423/open-gateway/internal/config"
	"github.com/c25423/open-gateway/internal/handler"
	"github.com/c25423/open-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
)

var (
	configFilePath string
	dataDirPath    string
)

func main() {
	pflag.StringVarP(&configFilePath, "config-file", "c", "/etc/open-gateway/config.yaml", "Path to configuration file.")
	pflag.StringVarP(&dataDirPath, "data-dir", "d", "/var/lib/open-gateway", "Path to served files.")
	pflag.Parse()
	log.Printf("Config file path: %s\n", configFilePath)
	log.Printf("Data dir path: %s\n", dataDirPath)

	if err := config.Load(configFilePath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	r := gin.Default()

	r.Use(middleware.BearerAuthMiddleware())
	r.GET("/models", handler.NewModelsHandler())
	r.POST("/chat/completions", handler.NewChatCompletionsHandler())

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.GetHost(), config.GetPort()),
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
