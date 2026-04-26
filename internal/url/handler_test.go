package url

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"generate-short-url/internal/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type URLServiceMock struct {
	mock.Mock
}

func (m *URLServiceMock) Create(req CreateShortURLRequest) (*ShortURLResponse, error) {
	args := m.Called(req)

	var result *ShortURLResponse
	if value := args.Get(0); value != nil {
		result = value.(*ShortURLResponse)
	}

	return result, args.Error(1)
}

func (m *URLServiceMock) GetAll(isActive *bool) ([]ShortURLResponse, error) {
	args := m.Called(isActive)

	var urls []ShortURLResponse
	if value := args.Get(0); value != nil {
		urls = value.([]ShortURLResponse)
	}

	return urls, args.Error(1)
}

func (m *URLServiceMock) GetByCode(shortCode string) (*ShortURLResponse, error) {
	args := m.Called(shortCode)

	var result *ShortURLResponse
	if value := args.Get(0); value != nil {
		result = value.(*ShortURLResponse)
	}

	return result, args.Error(1)
}

func (m *URLServiceMock) Desactive(shortCode string) (string, error) {
	args := m.Called(shortCode)

	var result string
	if value := args.Get(0); value != nil {
		result = value.(string)
	}

	return result, args.Error(1)
}

func (m *URLServiceMock) Active(shortCode string) (string, error) {
	args := m.Called(shortCode)

	var result string
	if value := args.Get(0); value != nil {
		result = value.(string)
	}

	return result, args.Error(1)
}

func (m *URLServiceMock) Redirect(shortCode string) (string, error) {
	args := m.Called(shortCode)

	var result string
	if value := args.Get(0); value != nil {
		result = value.(string)
	}

	return result, args.Error(1)
}

func setupHandlerRouter(service URLService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middlewares.ErrorHandler())

	handler := NewHandlerUrl(service)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/urls", handler.Create).
			GET("/urls", handler.GetAll).
			GET("/urls/:code", handler.GetByCode).
			PATCH("/urls/:code/deactivate", handler.Desactive).
			PATCH("/urls/:code/desactive", handler.Desactive).
			PATCH("/urls/:code/active", handler.Active)
	}
	r.GET("/:code", handler.Redirect)

	return r
}

func performRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func TestHandlerUrl_Create(t *testing.T) {
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	testCases := []struct {
		name       string
		body       string
		setupMock  func(*URLServiceMock)
		statusCode int
		expected   string
	}{
		{
			name: "success",
			body: `{"originalUrl":"https://google.com"}`,
			setupMock: func(service *URLServiceMock) {
				service.On("Create", CreateShortURLRequest{OriginalURL: "https://google.com"}).
					Return(&ShortURLResponse{ShortURL: "http://sho.rt/abc123"}, nil)
			},
			statusCode: http.StatusOK,
			expected:   `{"shortUrl":"http://sho.rt/abc123"}`,
		},
		{
			name:       "invalid payload",
			body:       `{}`,
			statusCode: http.StatusBadRequest,
			expected:   `{"success":false,"error":{"code":"BAD_REQUEST","message":"Invalid request payload. originalUrl is required and expiresAt must be a valid datetime (RFC3339)"}}`,
		},
		{
			name:       "invalid original url",
			body:       `{"originalUrl":"ftp://google.com"}`,
			statusCode: http.StatusBadRequest,
			expected:   `{"success":false,"error":{"code":"BAD_REQUEST","message":"Original url must be a valid http or https URL"}}`,
		},
		{
			name:       "expires at in past",
			body:       fmt.Sprintf(`{"originalUrl":"https://google.com","expiresAt":"%s"}`, past),
			statusCode: http.StatusBadRequest,
			expected:   `{"success":false,"error":{"code":"BAD_REQUEST","message":"expiresAt must be a future datetime"}}`,
		},
		{
			name: "service error",
			body: `{"originalUrl":"https://google.com"}`,
			setupMock: func(service *URLServiceMock) {
				service.On("Create", CreateShortURLRequest{OriginalURL: "https://google.com"}).
					Return((*ShortURLResponse)(nil), middlewares.ErrBadRequest)
			},
			statusCode: http.StatusBadRequest,
			expected:   `{"success":false,"error":{"code":"BAD_REQUEST","message":"invalid request"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := new(URLServiceMock)
			if tc.setupMock != nil {
				tc.setupMock(service)
			}

			router := setupHandlerRouter(service)
			w := performRequest(router, http.MethodPost, "/api/v1/urls", tc.body)

			assert.Equal(t, tc.statusCode, w.Code)
			assert.JSONEq(t, tc.expected, w.Body.String())
			service.AssertExpectations(t)
		})
	}
}

func TestHandlerUrl_GetAll(t *testing.T) {
	testCases := []struct {
		name       string
		path       string
		setupMock  func(*URLServiceMock)
		statusCode int
		expected   string
	}{
		{
			name: "success without filter",
			path: "/api/v1/urls",
			setupMock: func(service *URLServiceMock) {
				service.On("GetAll", (*bool)(nil)).
					Return([]ShortURLResponse{{ShortCode: "abc123", OriginalURL: "https://google.com"}}, nil)
			},
			statusCode: http.StatusOK,
			expected:   `[{"shortCode":"abc123","originalUrl":"https://google.com","id":"00000000-0000-0000-0000-000000000000","shortUrl":"","clickCount":0,"createdAt":"0001-01-01T00:00:00Z","expiresAt":null,"isActive":false}]`,
		},
		{
			name: "success with filter",
			path: "/api/v1/urls?is_active=true",
			setupMock: func(service *URLServiceMock) {
				service.On("GetAll", mock.MatchedBy(func(isActive *bool) bool {
					return isActive != nil && *isActive
				})).Return([]ShortURLResponse{{ShortCode: "abc123", IsActive: true}}, nil)
			},
			statusCode: http.StatusOK,
			expected:   `[{"shortCode":"abc123","id":"00000000-0000-0000-0000-000000000000","originalUrl":"","shortUrl":"","clickCount":0,"createdAt":"0001-01-01T00:00:00Z","expiresAt":null,"isActive":true}]`,
		},
		{
			name:       "invalid filter",
			path:       "/api/v1/urls?is_active=invalid",
			statusCode: http.StatusBadRequest,
			expected:   `{"success":false,"error":{"code":"BAD_REQUEST","message":"is_active must be true or false"}}`,
		},
		{
			name: "service error",
			path: "/api/v1/urls",
			setupMock: func(service *URLServiceMock) {
				service.On("GetAll", (*bool)(nil)).Return(([]ShortURLResponse)(nil), middlewares.ErrBadRequest)
			},
			statusCode: http.StatusBadRequest,
			expected:   `{"success":false,"error":{"code":"BAD_REQUEST","message":"invalid request"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := new(URLServiceMock)
			if tc.setupMock != nil {
				tc.setupMock(service)
			}

			router := setupHandlerRouter(service)
			w := performRequest(router, http.MethodGet, tc.path, "")

			assert.Equal(t, tc.statusCode, w.Code)
			assert.JSONEq(t, tc.expected, w.Body.String())
			service.AssertExpectations(t)
		})
	}
}

func TestHandlerUrl_GetByCode(t *testing.T) {
	testCases := []struct {
		name       string
		setupMock  func(*URLServiceMock)
		statusCode int
		expected   string
	}{
		{
			name: "success",
			setupMock: func(service *URLServiceMock) {
				service.On("GetByCode", "abc123").
					Return(&ShortURLResponse{ShortCode: "abc123", OriginalURL: "https://google.com"}, nil)
			},
			statusCode: http.StatusOK,
			expected:   `{"shortCode":"abc123","originalUrl":"https://google.com","id":"00000000-0000-0000-0000-000000000000","shortUrl":"","clickCount":0,"createdAt":"0001-01-01T00:00:00Z","expiresAt":null,"isActive":false}`,
		},
		{
			name: "not found",
			setupMock: func(service *URLServiceMock) {
				service.On("GetByCode", "abc123").Return((*ShortURLResponse)(nil), middlewares.ErrNotFound)
			},
			statusCode: http.StatusNotFound,
			expected:   `{"success":false,"error":{"code":"NOT_FOUND","message":"resource not found"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := new(URLServiceMock)
			tc.setupMock(service)

			router := setupHandlerRouter(service)
			w := performRequest(router, http.MethodGet, "/api/v1/urls/abc123", "")

			assert.Equal(t, tc.statusCode, w.Code)
			assert.JSONEq(t, tc.expected, w.Body.String())
			service.AssertExpectations(t)
		})
	}
}

func TestHandlerUrl_Desactive(t *testing.T) {
	testCases := []struct {
		name       string
		setupMock  func(*URLServiceMock)
		statusCode int
		expected   string
	}{
		{
			name: "success",
			setupMock: func(service *URLServiceMock) {
				service.On("Desactive", "abc123").Return("short url successfully disabled", nil)
			},
			statusCode: http.StatusOK,
			expected:   `{"success":true,"message":"short url successfully disabled"}`,
		},
		{
			name: "not found",
			setupMock: func(service *URLServiceMock) {
				service.On("Desactive", "abc123").Return("", middlewares.ErrNotFound)
			},
			statusCode: http.StatusNotFound,
			expected:   `{"success":false,"error":{"code":"NOT_FOUND","message":"resource not found"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := new(URLServiceMock)
			tc.setupMock(service)

			router := setupHandlerRouter(service)
			w := performRequest(router, http.MethodPatch, "/api/v1/urls/abc123/deactivate", "")

			assert.Equal(t, tc.statusCode, w.Code)
			assert.JSONEq(t, tc.expected, w.Body.String())
			service.AssertExpectations(t)
		})
	}
}

func TestHandlerUrl_Active(t *testing.T) {
	testCases := []struct {
		name       string
		setupMock  func(*URLServiceMock)
		statusCode int
		expected   string
	}{
		{
			name: "success",
			setupMock: func(service *URLServiceMock) {
				service.On("Active", "abc123").Return("short url successfully activated", nil)
			},
			statusCode: http.StatusOK,
			expected:   `{"success":true,"message":"short url successfully activated"}`,
		},
		{
			name: "not found",
			setupMock: func(service *URLServiceMock) {
				service.On("Active", "abc123").Return("", middlewares.ErrNotFound)
			},
			statusCode: http.StatusNotFound,
			expected:   `{"success":false,"error":{"code":"NOT_FOUND","message":"resource not found"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := new(URLServiceMock)
			tc.setupMock(service)

			router := setupHandlerRouter(service)
			w := performRequest(router, http.MethodPatch, "/api/v1/urls/abc123/active", "")

			assert.Equal(t, tc.statusCode, w.Code)
			assert.JSONEq(t, tc.expected, w.Body.String())
			service.AssertExpectations(t)
		})
	}
}

func TestHandlerUrl_Redirect(t *testing.T) {
	testCases := []struct {
		name       string
		setupMock  func(*URLServiceMock)
		statusCode int
		expected   string
	}{
		{
			name: "success",
			setupMock: func(service *URLServiceMock) {
				service.On("Redirect", "abc123").Return("https://google.com", nil)
			},
			statusCode: http.StatusFound,
		},
		{
			name: "service error",
			setupMock: func(service *URLServiceMock) {
				service.On("Redirect", "abc123").Return("", middlewares.ErrExpiredUrl)
			},
			statusCode: http.StatusBadRequest,
			expected:   `{"success":false,"error":{"code":"URL_EXPIRED","message":"short url is expired"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := new(URLServiceMock)
			tc.setupMock(service)

			router := setupHandlerRouter(service)
			w := performRequest(router, http.MethodGet, "/abc123", "")

			assert.Equal(t, tc.statusCode, w.Code)

			if tc.statusCode == http.StatusFound {
				assert.Equal(t, "https://google.com", w.Header().Get("Location"))
				assert.Equal(t, "no-store, no-cache, must-revalidate, max-age=0", w.Header().Get("Cache-Control"))
				assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
				assert.Equal(t, "0", w.Header().Get("Expires"))
			} else {
				assert.JSONEq(t, tc.expected, w.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
