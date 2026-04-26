package url

import (
	"errors"
	"generate-short-url/internal/middlewares"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceUrl_Create_Success(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	request := CreateShortURLRequest{
		OriginalURL: "https://www.original.com",
	}

	repoMock.On("Create", mock.AnythingOfType("CreateShortURLRequest")).
		Return(ShortUrl{
			ID:          uuid.New(),
			OriginalURL: "https://www.original.com",
			ShortCode:   "abc123",
			ClickCount:  0,
			IsActive:    true,
		}, nil)

	result, err := service.Create(request)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "https://www.original.com", result.OriginalURL)
	assert.Equal(t, "abc123", result.ShortCode)
	assert.Len(t, result.ShortCode, 6)
	assert.True(t, result.IsActive)
	assert.Equal(t, 0, result.ClickCount)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_Create_BadRequest(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	request := CreateShortURLRequest{
		OriginalURL: "https://www.original.com",
	}

	expectedErr := middlewares.ErrBadRequest

	repoMock.On("Create", request).Return(ShortUrl{}, expectedErr)

	result, err := service.Create(request)

	assert.ErrorIs(t, expectedErr, err)
	assert.Nil(t, result)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_GetAll_Success(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	isActive := true

	urls := []ShortUrl{
		{
			OriginalURL: "https://google.com",
			ShortCode:   "abc123",
			IsActive:    true,
		},
		{
			OriginalURL: "https://github.com",
			ShortCode:   "xyz789",
			IsActive:    true,
		},
	}

	repoMock.On("GetAll", &isActive).Return(urls, nil)

	result, err := service.GetAll(&isActive)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	assert.Equal(t, "abc123", result[0].ShortCode)
	assert.Equal(t, "https://google.com", result[0].OriginalURL)
	assert.Equal(t, true, result[0].IsActive)

	assert.Equal(t, "xyz789", result[1].ShortCode)
	assert.Equal(t, "https://github.com", result[1].OriginalURL)
	assert.Equal(t, true, result[1].IsActive)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_GetAll_Error(t *testing.T) {
	repoMock := new(RepositoryMock)

	service := ServiceUrl{
		repo: repoMock,
	}

	isActive := true
	expectedErr := errors.New("database error")

	repoMock.On("GetAll", &isActive).Return(nil, expectedErr)

	result, err := service.GetAll(&isActive)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_GetAll_NoFilter(t *testing.T) {
	repoMock := new(RepositoryMock)

	service := ServiceUrl{
		repo: repoMock,
	}

	urls := []ShortUrl{
		{
			OriginalURL: "https://google.com",
			ShortCode:   "abc123",
			IsActive:    true,
		},
		{
			OriginalURL: "https://inactive.com",
			ShortCode:   "def456",
			IsActive:    false,
		},
	}

	repoMock.On("GetAll", (*bool)(nil)).Return(urls, nil)

	result, err := service.GetAll(nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	assert.Equal(t, "abc123", result[0].ShortCode)
	assert.Equal(t, "def456", result[1].ShortCode)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_GetByCode_Success(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	url := ShortUrl{
		ID:          uuid.New(),
		OriginalURL: "https://github.com",
		ShortCode:   "xyz789",
		IsActive:    true,
	}

	repoMock.On("GetByCode", mock.AnythingOfType("string")).
		Return(url, nil)

	result, err := service.GetByCode(url.ShortCode)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "xyz789", result.ShortCode)
	assert.Equal(t, "https://github.com", result.OriginalURL)
	assert.Equal(t, true, result.IsActive)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_GetByCode_NotFound(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	url := ShortUrl{
		ShortCode: "xyz789",
	}

	errorExpected := middlewares.ErrNotFound

	repoMock.On("GetByCode", mock.AnythingOfType("string")).
		Return(ShortUrl{}, errorExpected)

	result, err := service.GetByCode(url.ShortCode)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, errorExpected)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_Desactive_Success(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	shortcode := "abc123"

	repoMock.On("Desactive", mock.AnythingOfType("string")).
		Return("short url successfully disabled", nil)

	result, err := service.Desactive(shortcode)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "short url successfully disabled", result)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_Desactive_NotFound(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	shortcode := "cba123"
	expectedErr := middlewares.ErrNotFound

	repoMock.On("Desactive", mock.AnythingOfType("string")).
		Return("", expectedErr)

	result, err := service.Desactive(shortcode)

	assert.Error(t, err)
	assert.Empty(t, result)

	assert.Equal(t, expectedErr, err)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_Active_Success(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	shortcode := "abc123"

	repoMock.On("Active", mock.AnythingOfType("string")).
		Return("short url successfully activated", nil)

	result, err := service.Active(shortcode)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, "short url successfully activated", result)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_Active_NotFound(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	shortcode := "cba123"
	expectedErr := middlewares.ErrNotFound

	repoMock.On("Active", mock.AnythingOfType("string")).
		Return("", expectedErr)

	result, err := service.Active(shortcode)

	assert.Error(t, err)
	assert.Empty(t, result)

	assert.Equal(t, expectedErr, err)

	repoMock.AssertExpectations(t)
}

func TestServiceUrl_Redirect_Success(t *testing.T) {
	repoMock := new(RepositoryMock)
	service := NewServiceUrl(repoMock)

	shortcode := "abc123"
	originalURL := "https://www.original.com"

	repoMock.On("Redirect", mock.AnythingOfType("string")).
		Return(originalURL, nil)

	result, err := service.Redirect(shortcode)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, originalURL, result)

	repoMock.AssertExpectations(t)
}
