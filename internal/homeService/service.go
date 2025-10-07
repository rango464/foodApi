package homeService

type HomeService interface {
	Home() error
}

type homeService struct {
	repo HomeRepository
}

func NewHomeService(repo HomeRepository) HomeService {
	return &homeService{repo: repo}
}

func (s *homeService) Home() error { // проверка валидности email
	return s.repo.Home()

}
