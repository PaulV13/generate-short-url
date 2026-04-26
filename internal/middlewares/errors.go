package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrNotFound           = &AppError{Status: 404, Code: "NOT_FOUND", Message: "resource not found"}
	ErrUnauthorized       = &AppError{Status: 401, Code: "UNAUTHORIZED", Message: "authentication required"}
	ErrForbidden          = &AppError{Status: 403, Code: "FORBIDDEN", Message: "not have permission to access"}
	ErrBadRequest         = &AppError{Status: 400, Code: "BAD_REQUEST", Message: "invalid request"}
	ErrDuplicatedKey      = &AppError{Status: 400, Code: "DUPLICATED_KEY", Message: "record duplicated key"}
	ErrForeignKeyViolated = &AppError{Status: 400, Code: "FOREIGN_KEY_VIOLATED", Message: "foreign key violated"}
	ErrExpiredUrl         = &AppError{Status: 400, Code: "URL_EXPIRED", Message: "short url is expired"}
	ErrUrlNotActive       = &AppError{Status: 400, Code: "URL_NOT_ACTIVE", Message: "short url is not active"}
)

// ErrorHandler is a middleware that catches errors set via c.Error().
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		var appErr *AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Status, gin.H{
				"success": false,
				"error":   gin.H{"code": appErr.Code, "message": appErr.Message},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   gin.H{"code": "INTERNAL", "message": "an unexpected error occurred"},
			})
		}
	}
}
