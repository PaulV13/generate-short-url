package url

import (
	"errors"
	"time"

	"generate-short-url/internal/middlewares"
	"generate-short-url/utils/shortcode"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository interface {
	Create(req CreateShortURLRequest) (ShortUrl, error)
	GetAll(isActive *bool) ([]ShortUrl, error)
	GetByCode(shortCode string) (ShortUrl, error)
	Desactive(shortCode string) (string, error)
	Active(shortCode string) (string, error)
	Redirect(shortCode string) (string, error)
}

type RepositoryUrl struct {
	db           *gorm.DB
	generateCode func() string
}

func NewRepositoryUrl(db *gorm.DB) *RepositoryUrl {
	return &RepositoryUrl{
		db:           db,
		generateCode: shortcode.GenerateShortURL,
	}
}

func (r *RepositoryUrl) Create(req CreateShortURLRequest) (ShortUrl, error) {
	const maxCreateRetries = 5

	for i := 0; i < maxCreateRetries; i++ {
		shortCode := r.generateCode()
		if shortCode == "" {
			return ShortUrl{}, middlewares.ErrBadRequest
		}

		url := ShortUrl{
			OriginalURL: req.OriginalURL,
			ShortCode:   shortCode,
			CreatedAt:   time.Now(),
			ExpiresAt:   req.ExpiresAt,
		}

		if err := r.db.Create(&url).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			return ShortUrl{}, middlewares.ErrBadRequest
		}

		return url, nil
	}

	return ShortUrl{}, middlewares.ErrDuplicatedKey
}

func (r *RepositoryUrl) GetAll(isActive *bool) ([]ShortUrl, error) {
	var urls []ShortUrl

	query := r.db.Model(&ShortUrl{})

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	err := query.Find(&urls).Error
	if err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *RepositoryUrl) GetByCode(shortCode string) (ShortUrl, error) {
	var shortUrl ShortUrl
	if err := r.db.Where("short_code = ?", shortCode).First(&shortUrl).Error; err != nil {
		return ShortUrl{}, middlewares.ErrNotFound
	}

	return shortUrl, nil
}

func (r *RepositoryUrl) Desactive(shortCode string) (string, error) {
	var shortUrl ShortUrl
	if err := r.db.Where("short_code = ?", shortCode).First(&shortUrl).Error; err != nil {
		return "", middlewares.ErrNotFound
	}

	if err := r.db.Model(&ShortUrl{}).
		Where("ID = ?", shortUrl.ID).
		Update("is_active", false).Error; err != nil {
		return "", middlewares.ErrBadRequest
	}

	return "short url successfully disabled", nil
}

func (r *RepositoryUrl) Active(shortCode string) (string, error) {
	var shortUrl ShortUrl
	if err := r.db.Where("short_code = ?", shortCode).First(&shortUrl).Error; err != nil {
		return "", middlewares.ErrNotFound
	}

	if err := r.db.Model(&ShortUrl{}).
		Where("ID = ?", shortUrl.ID).
		Update("is_active", true).Error; err != nil {
		return "", middlewares.ErrBadRequest
	}

	return "short url successfully activated", nil
}

func (r *RepositoryUrl) Redirect(shortCode string) (string, error) {
	var shortUrl ShortUrl

	if err := r.db.Where("short_code = ?", shortCode).First(&shortUrl).Error; err != nil {
		return "", middlewares.ErrNotFound
	}

	if shortUrl.ExpiresAt != nil && shortUrl.ExpiresAt.Before(time.Now()) {
		return "", middlewares.ErrExpiredUrl
	}

	if shortUrl.IsActive != true {
		return "", middlewares.ErrUrlNotActive
	}

	if err := r.db.Model(&ShortUrl{}).
		Where("id = ?", shortUrl.ID).
		Update("click_count", shortUrl.ClickCount+1).Error; err != nil {
		return "", middlewares.ErrBadRequest
	}

	return shortUrl.OriginalURL, nil
}
