package url

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type HandlerUrl struct {
	service URLService
}

type URLService interface {
	Create(req CreateShortURLRequest) (*ShortURLResponse, error)
	GetAll(isActive *bool) ([]ShortURLResponse, error)
	GetByCode(shortCode string) (*ShortURLResponse, error)
	Desactive(shortCode string) (string, error)
	Active(shortCode string) (string, error)
	Redirect(shortCode string) (string, error)
}

func NewHandlerUrl(service URLService) *HandlerUrl {
	return &HandlerUrl{
		service: service,
	}
}

func (h *HandlerUrl) Create(c *gin.Context) {
	var req CreateShortURLRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Invalid request payload. originalUrl is required and expiresAt must be a valid datetime (RFC3339)"},
		})
		return
	}

	parsedURL, err := url.ParseRequestURI(req.OriginalURL)
	if err != nil || parsedURL == nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "Original url must be a valid http or https URL"},
		})
		return
	}

	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "expiresAt must be a future datetime"},
		})
		return
	}

	result, err := h.service.Create(req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"shortUrl": result.ShortURL})
}

func (h *HandlerUrl) GetAll(c *gin.Context) {
	isActiveStr := c.Query("is_active")

	var isActive *bool

	if isActiveStr != "" {
		parsed, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"code": "BAD_REQUEST", "message": "is_active must be true or false"},
			})
			return
		}
		isActive = &parsed
	}

	urls, err := h.service.GetAll(isActive)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, urls)
}

func (h *HandlerUrl) GetByCode(c *gin.Context) {
	shortCode := c.Param("code")

	url, err := h.service.GetByCode(shortCode)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, url)
}

func (h *HandlerUrl) Desactive(c *gin.Context) {
	shortCode := c.Param("code")

	message, err := h.service.Desactive(shortCode)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": message})
}

func (h *HandlerUrl) Active(c *gin.Context) {
	shortCode := c.Param("code")

	message, err := h.service.Active(shortCode)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": message})
}

func (h *HandlerUrl) Redirect(c *gin.Context) {
	shortCode := c.Param("code")

	message, err := h.service.Redirect(shortCode)
	if err != nil {
		c.Error(err)
		return
	}

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.Redirect(http.StatusFound, message)
}
