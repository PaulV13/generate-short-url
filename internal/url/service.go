package url

import "os"

type ServiceUrl struct {
	repo    Repository
	baseURL string
}

var _ URLService = (*ServiceUrl)(nil)

func NewServiceUrl(repo Repository) *ServiceUrl {
	return &ServiceUrl{
		repo:    repo,
		baseURL: os.Getenv("BASE_URL"),
	}
}

func (s *ServiceUrl) Create(req CreateShortURLRequest) (*ShortURLResponse, error) {
	url, err := s.repo.Create(req)
	if err != nil {
		return nil, err
	}

	return MapperShortUrl(url, s.baseURL), err
}

func (s *ServiceUrl) GetAll(isActive *bool) ([]ShortURLResponse, error) {
	urls, err := s.repo.GetAll(isActive)
	if err != nil {
		return nil, err
	}

	baseURL := os.Getenv("BASE_URL")

	return MapperShortUrlList(urls, baseURL), nil
}

func (s *ServiceUrl) GetByCode(shortCode string) (*ShortURLResponse, error) {
	url, err := s.repo.GetByCode(shortCode)
	if err != nil {
		return nil, err
	}

	return MapperShortUrl(url, s.baseURL), nil
}

func (s *ServiceUrl) Desactive(shortCode string) (string, error) {
	return s.repo.Desactive(shortCode)
}

func (s *ServiceUrl) Active(shortCode string) (string, error) {
	return s.repo.Active(shortCode)
}

func (s *ServiceUrl) Redirect(shortCode string) (string, error) {
	return s.repo.Redirect(shortCode)
}
