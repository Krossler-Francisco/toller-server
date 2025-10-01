package users

type UserService struct {
	Repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{Repo: repo}
}

func (s *UserService) GetAllUsers() ([]User, error) {
	return s.Repo.GetAllUsers()
}

func (s *UserService) GetUserByID(id int) (*User, error) {
	return s.Repo.GetUserByID(id)
}

func (s *UserService) SearchUsers(query string) ([]User, error) {
	return s.Repo.SearchUsers(query)
}
