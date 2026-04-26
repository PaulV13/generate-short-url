package url

import (
	"time"

	"github.com/google/uuid"
)

type ShortUrl struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OriginalURL string    `gorm:"not null"`
	ShortCode   string    `gorm:"size:20;not null;uniqueIndex"`
	ClickCount  int       `gorm:"not null;default:0"`
	CreatedAt   time.Time `gorm:"not null"`
	ExpiresAt   *time.Time
	IsActive    bool `gorm:"not null;default:true"`
}

type CreateShortURLRequest struct {
	OriginalURL string     `json:"originalUrl" binding:"required"`
	ExpiresAt   *time.Time `json:"expiresAt"`
}

type ShortURLResponse struct {
	ID          uuid.UUID  `json:"id"`
	OriginalURL string     `json:"originalUrl"`
	ShortCode   string     `json:"shortCode"`
	ShortURL    string     `json:"shortUrl"`
	ClickCount  int        `json:"clickCount"`
	CreatedAt   time.Time  `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	IsActive    bool       `json:"isActive"`
}

func MapperShortUrl(url ShortUrl, baseURL string) *ShortURLResponse {
	urlResponse := ShortURLResponse{
		ID:          url.ID,
		OriginalURL: url.OriginalURL,
		ShortCode:   url.ShortCode,
		ShortURL:    baseURL + url.ShortCode,
		ClickCount:  url.ClickCount,
		IsActive:    url.IsActive,
		CreatedAt:   url.CreatedAt,
		ExpiresAt:   url.ExpiresAt,
	}

	return &urlResponse
}

func MapperShortUrlList(urls []ShortUrl, baseURL string) []ShortURLResponse {
	urlsResponse := make([]ShortURLResponse, 0, len(urls))

	for _, u := range urls {
		urlsResponse = append(urlsResponse, ShortURLResponse{
			ID:          u.ID,
			OriginalURL: u.OriginalURL,
			ShortCode:   u.ShortCode,
			ShortURL:    baseURL + u.ShortCode,
			ClickCount:  u.ClickCount,
			IsActive:    u.IsActive,
			CreatedAt:   u.CreatedAt,
			ExpiresAt:   u.ExpiresAt,
		})
	}

	return urlsResponse
}
