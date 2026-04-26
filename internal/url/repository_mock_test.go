package url

import "github.com/stretchr/testify/mock"

type RepositoryMock struct {
	mock.Mock
}

func (m *RepositoryMock) Create(req CreateShortURLRequest) (ShortUrl, error) {
	args := m.Called(req)
	return args.Get(0).(ShortUrl), args.Error(1)
}

func (m *RepositoryMock) GetAll(isActive *bool) ([]ShortUrl, error) {
	args := m.Called(isActive)

	var urls []ShortUrl
	if value := args.Get(0); value != nil {
		urls = value.([]ShortUrl)
	}

	return urls, args.Error(1)
}

func (m *RepositoryMock) GetByCode(shortCode string) (ShortUrl, error) {
	args := m.Called(shortCode)

	var url ShortUrl
	if value := args.Get(0); value != nil {
		url = value.(ShortUrl)
	}

	return url, args.Error(1)
}

func (m *RepositoryMock) Desactive(shortCode string) (string, error) {
	args := m.Called(shortCode)

	var result string
	if value := args.Get(0); value != nil {
		result = value.(string)
	}

	return result, args.Error(1)
}

func (m *RepositoryMock) Active(shortCode string) (string, error) {
	args := m.Called(shortCode)

	var result string
	if value := args.Get(0); value != nil {
		result = value.(string)
	}

	return result, args.Error(1)
}

func (m *RepositoryMock) Redirect(shortCode string) (string, error) {
	args := m.Called(shortCode)

	var result string
	if value := args.Get(0); value != nil {
		result = value.(string)
	}

	return result, args.Error(1)
}
