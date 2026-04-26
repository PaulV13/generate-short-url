package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"generate-short-url/internal/config"
	"generate-short-url/internal/db"
	"generate-short-url/internal/middlewares"
	"generate-short-url/internal/url"

	"github.com/gin-gonic/gin"
)

func Start() {
	if err := config.ValidateRequiredEnv(); err != nil {
		log.Fatal(err)
	}

	database := db.DatabaseConnection()
	if err := db.HealthCheck(database); err != nil {
		log.Fatal("database health check failed: ", err)
	}
	defer func() {
		if err := db.Close(database); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	r := gin.New()
	r.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/health"}}),
		gin.Recovery(),
		middlewares.ErrorHandler(),
	)
	r.GET("/health", func(c *gin.Context) {
		if err := db.HealthCheck(database); err != nil {
			log.Printf("healthcheck failed: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	repoUrl := url.NewRepositoryUrl(database)
	serviceUrl := url.NewServiceUrl(repoUrl)
	handlerUrl := url.NewHandlerUrl(serviceUrl)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/urls", handlerUrl.Create).
			GET("/urls", handlerUrl.GetAll).
			GET("/urls/:code", handlerUrl.GetByCode).
			PATCH("/urls/:code/deactivate", handlerUrl.Desactive).
			PATCH("/urls/:code/active", handlerUrl.Active)
	}
	r.GET("/:code", handlerUrl.Redirect)

	port := os.Getenv("PORT")
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("starting server on port=%s base_url=%s", port, os.Getenv("BASE_URL"))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Print("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	log.Print("server stopped")
}
