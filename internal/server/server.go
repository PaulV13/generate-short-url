package server

import (
	"log"
	"net/http"
	"os"

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
		log.Fatal("database health check failed")
	}

	r := gin.Default()
	r.Use(middlewares.ErrorHandler())
	r.GET("/health", func(c *gin.Context) {
		if err := db.HealthCheck(database); err != nil {
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

	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}
}
