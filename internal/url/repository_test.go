package url

import (
	"generate-short-url/internal/middlewares"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var shortURLColumns = []string{"id", "original_url", "short_code", "click_count", "created_at", "expires_at", "is_active"}

var selectByShortCodeRegex = regexp.QuoteMeta(
	`SELECT * FROM "short_urls" WHERE short_code = $1 ORDER BY "short_urls"."id" LIMIT $2`,
)

func newShortURLRows(urls ...ShortUrl) *sqlmock.Rows {
	rows := sqlmock.NewRows(shortURLColumns)
	for _, u := range urls {
		rows.AddRow(u.ID, u.OriginalURL, u.ShortCode, u.ClickCount, u.CreatedAt, u.ExpiresAt, u.IsActive)
	}

	return rows
}

func expectIsActiveUpdate(mock sqlmock.Sqlmock, value bool, id uuid.UUID, updateErr error) {
	mock.ExpectBegin()
	exec := mock.ExpectExec(regexp.QuoteMeta(`UPDATE "short_urls" SET "is_active"=$1 WHERE ID = $2`)).WithArgs(value, id)

	if updateErr != nil {
		exec.WillReturnError(updateErr)
		mock.ExpectRollback()
		return
	}

	exec.WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func setupRepositoryTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = mockDB.Close()
	})

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: mockDB,
	}), &gorm.Config{})

	assert.NoError(t, err)

	return db, mock
}

func TestRepository_Create_Success(t *testing.T) {
	db, mock := setupRepositoryTest(t)

	repo := NewRepositoryUrl(db)
	repo.generateCode = func() string {
		return "abc123"
	}

	req := CreateShortURLRequest{
		OriginalURL: "https://google.com",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO "short_urls"`).
		WithArgs(
			req.OriginalURL,
			"abc123",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	mock.ExpectCommit()

	result, err := repo.Create(req)

	assert.NoError(t, err)
	assert.Equal(t, req.OriginalURL, result.OriginalURL)
	assert.Equal(t, "abc123", result.ShortCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryUrl_Create_EmptyCode_ReturnsBadRequest(t *testing.T) {
	db, _ := setupRepositoryTest(t)

	repo := NewRepositoryUrl(db)
	repo.generateCode = func() string {
		return ""
	}

	req := CreateShortURLRequest{
		OriginalURL: "https://google.com",
	}

	result, err := repo.Create(req)

	assert.ErrorIs(t, err, middlewares.ErrBadRequest)
	assert.Equal(t, ShortUrl{}, result)
}

func TestRepositoryUrl_Create_Duplicate(t *testing.T) {
	db, mock := setupRepositoryTest(t)

	repo := NewRepositoryUrl(db)
	repo.generateCode = func() string {
		return "abc123"
	}

	req := CreateShortURLRequest{
		OriginalURL: "https://google.com",
	}

	insertRegex := regexp.QuoteMeta(
		`INSERT INTO "short_urls" ("original_url","short_code","click_count","created_at","expires_at","is_active") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`,
	)

	for i := 0; i < 5; i++ {
		mock.ExpectBegin()
		mock.ExpectQuery(insertRegex).
			WithArgs(
				req.OriginalURL,
				"abc123",
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(&pgconn.PgError{Code: "23505"})
		mock.ExpectRollback()
	}

	result, err := repo.Create(req)

	assert.Error(t, err)
	assert.Equal(t, ShortUrl{}, result)
	assert.Equal(t, middlewares.ErrDuplicatedKey, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryUrl_Create_DuplicateThenSuccess(t *testing.T) {
	db, mock := setupRepositoryTest(t)

	codes := []string{"duplicado", "abc123"}
	i := 0

	repo := NewRepositoryUrl(db)
	repo.generateCode = func() string {
		code := codes[i]
		i++
		return code
	}

	req := CreateShortURLRequest{
		OriginalURL: "https://google.com",
	}

	insertRegex := regexp.QuoteMeta(
		`INSERT INTO "short_urls" ("original_url","short_code","click_count","created_at","expires_at","is_active") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`,
	)

	mock.ExpectBegin()
	mock.ExpectQuery(insertRegex).
		WithArgs(
			req.OriginalURL,
			"duplicado",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(&pgconn.PgError{Code: "23505"})
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(insertRegex).
		WithArgs(
			req.OriginalURL,
			"abc123",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	result, err := repo.Create(req)

	assert.NoError(t, err)
	assert.Equal(t, "abc123", result.ShortCode)
	assert.Equal(t, req.OriginalURL, result.OriginalURL)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryUrl_GetAll(t *testing.T) {
	testCases := []struct {
		name        string
		hasFilter   bool
		filter      bool
		query       string
		rows        *sqlmock.Rows
		queryErr    error
		wantErr     error
		wantLen     int
		wantIsNil   bool
		wantFirstSC string
	}{
		{
			name:        "success without filter",
			query:       regexp.QuoteMeta(`SELECT * FROM "short_urls"`),
			rows:        newShortURLRows(ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", ClickCount: 3, CreatedAt: time.Now(), IsActive: true}, ShortUrl{ID: uuid.New(), OriginalURL: "https://github.com", ShortCode: "xyz789", ClickCount: 2, CreatedAt: time.Now(), IsActive: false}),
			wantLen:     2,
			wantFirstSC: "abc123",
		},
		{
			name:        "success with active filter",
			hasFilter:   true,
			filter:      true,
			query:       regexp.QuoteMeta(`SELECT * FROM "short_urls" WHERE is_active = $1`),
			rows:        newShortURLRows(ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", ClickCount: 3, CreatedAt: time.Now(), IsActive: true}),
			wantLen:     1,
			wantFirstSC: "abc123",
		},
		{
			name:      "query error",
			query:     regexp.QuoteMeta(`SELECT * FROM "short_urls"`),
			queryErr:  assert.AnError,
			wantErr:   assert.AnError,
			wantIsNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupRepositoryTest(t)
			repo := NewRepositoryUrl(db)

			expectedQuery := mock.ExpectQuery(tc.query)
			var filterPtr *bool
			if tc.hasFilter {
				expectedQuery.WithArgs(tc.filter)
				filterPtr = &tc.filter
			}

			if tc.queryErr != nil {
				expectedQuery.WillReturnError(tc.queryErr)
			} else {
				expectedQuery.WillReturnRows(tc.rows)
			}

			urls, err := repo.GetAll(filterPtr)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, urls, tc.wantLen)
				if tc.wantLen > 0 {
					assert.Equal(t, tc.wantFirstSC, urls[0].ShortCode)
				}
			}

			if tc.wantIsNil {
				assert.Nil(t, urls)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryUrl_GetByCode(t *testing.T) {
	testCases := []struct {
		name      string
		shortCode string
		queryErr  error
		url       ShortUrl
		wantErr   error
	}{
		{
			name:      "success",
			shortCode: "abc123",
			url:       ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", CreatedAt: time.Now(), IsActive: true},
		},
		{
			name:      "not found",
			shortCode: "missing123",
			queryErr:  gorm.ErrRecordNotFound,
			wantErr:   middlewares.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupRepositoryTest(t)
			repo := NewRepositoryUrl(db)

			expectedQuery := mock.ExpectQuery(selectByShortCodeRegex).WithArgs(tc.shortCode, 1)
			if tc.queryErr != nil {
				expectedQuery.WillReturnError(tc.queryErr)
			} else {
				expectedQuery.WillReturnRows(newShortURLRows(tc.url))
			}

			result, err := repo.GetByCode(tc.shortCode)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, ShortUrl{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.url.ShortCode, result.ShortCode)
				assert.Equal(t, tc.url.OriginalURL, result.OriginalURL)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryUrl_Desactive(t *testing.T) {
	testCases := []struct {
		name        string
		shortCode   string
		selectErr   error
		selectedURL ShortUrl
		updateErr   error
		wantErr     error
		wantMessage string
	}{
		{
			name:        "success",
			shortCode:   "abc123",
			selectedURL: ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", CreatedAt: time.Now(), IsActive: true},
			wantMessage: "short url successfully disabled",
		},
		{
			name:      "not found",
			shortCode: "missing123",
			selectErr: gorm.ErrRecordNotFound,
			wantErr:   middlewares.ErrNotFound,
		},
		{
			name:        "update error",
			shortCode:   "abc123",
			selectedURL: ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", CreatedAt: time.Now(), IsActive: true},
			updateErr:   assert.AnError,
			wantErr:     middlewares.ErrBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupRepositoryTest(t)
			repo := NewRepositoryUrl(db)

			expectedQuery := mock.ExpectQuery(selectByShortCodeRegex).WithArgs(tc.shortCode, 1)
			if tc.selectErr != nil {
				expectedQuery.WillReturnError(tc.selectErr)
			} else {
				expectedQuery.WillReturnRows(newShortURLRows(tc.selectedURL))
				expectIsActiveUpdate(mock, false, tc.selectedURL.ID, tc.updateErr)
			}

			result, err := repo.Desactive(tc.shortCode)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantMessage, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryUrl_Active(t *testing.T) {
	testCases := []struct {
		name        string
		shortCode   string
		selectErr   error
		selectedURL ShortUrl
		updateErr   error
		wantErr     error
		wantMessage string
	}{
		{
			name:        "success",
			shortCode:   "abc123",
			selectedURL: ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", CreatedAt: time.Now(), IsActive: false},
			wantMessage: "short url successfully activated",
		},
		{
			name:      "not found",
			shortCode: "missing123",
			selectErr: gorm.ErrRecordNotFound,
			wantErr:   middlewares.ErrNotFound,
		},
		{
			name:        "update error",
			shortCode:   "abc123",
			selectedURL: ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", CreatedAt: time.Now(), IsActive: false},
			updateErr:   assert.AnError,
			wantErr:     middlewares.ErrBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupRepositoryTest(t)
			repo := NewRepositoryUrl(db)

			expectedQuery := mock.ExpectQuery(selectByShortCodeRegex).WithArgs(tc.shortCode, 1)
			if tc.selectErr != nil {
				expectedQuery.WillReturnError(tc.selectErr)
			} else {
				expectedQuery.WillReturnRows(newShortURLRows(tc.selectedURL))
				expectIsActiveUpdate(mock, true, tc.selectedURL.ID, tc.updateErr)
			}

			result, err := repo.Active(tc.shortCode)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantMessage, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepositoryUrl_Redirect(t *testing.T) {
	expiredAt := time.Now().Add(-time.Hour)
	testCases := []struct {
		name          string
		shortCode     string
		selectErr     error
		selectedURL   ShortUrl
		updateErr     error
		expectsUpdate bool
		wantErr       error
		wantResult    string
	}{
		{
			name:          "success",
			shortCode:     "abc123",
			selectedURL:   ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", CreatedAt: time.Now(), IsActive: true},
			expectsUpdate: true,
			wantResult:    "https://google.com",
		},
		{
			name:      "not found",
			shortCode: "missing123",
			selectErr: gorm.ErrRecordNotFound,
			wantErr:   middlewares.ErrNotFound,
		},
		{
			name:        "expired",
			shortCode:   "expired123",
			selectedURL: ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "expired123", CreatedAt: time.Now(), ExpiresAt: &expiredAt, IsActive: true},
			wantErr:     middlewares.ErrExpiredUrl,
		},
		{
			name:        "inactive",
			shortCode:   "inactive123",
			selectedURL: ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "inactive123", CreatedAt: time.Now(), IsActive: false},
			wantErr:     middlewares.ErrUrlNotActive,
		},
		{
			name:          "update click count error",
			shortCode:     "abc123",
			selectedURL:   ShortUrl{ID: uuid.New(), OriginalURL: "https://google.com", ShortCode: "abc123", CreatedAt: time.Now(), IsActive: true},
			updateErr:     assert.AnError,
			expectsUpdate: true,
			wantErr:       middlewares.ErrBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := setupRepositoryTest(t)
			repo := NewRepositoryUrl(db)

			expectedQuery := mock.ExpectQuery(selectByShortCodeRegex).WithArgs(tc.shortCode, 1)
			if tc.selectErr != nil {
				expectedQuery.WillReturnError(tc.selectErr)
			} else {
				expectedQuery.WillReturnRows(newShortURLRows(tc.selectedURL))
			}

			if tc.expectsUpdate {
				mock.ExpectBegin()
				exec := mock.ExpectExec(regexp.QuoteMeta(`UPDATE "short_urls" SET "click_count"=$1 WHERE id = $2`)).
					WithArgs(tc.selectedURL.ClickCount+1, tc.selectedURL.ID)

				if tc.updateErr != nil {
					exec.WillReturnError(tc.updateErr)
					mock.ExpectRollback()
				} else {
					exec.WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()
				}
			}

			result, err := repo.Redirect(tc.shortCode)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
