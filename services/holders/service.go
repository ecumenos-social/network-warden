package holders

type Repository interface{}

type Service interface{}

type service struct {
	repo Repository
}

func New(repo Repository) Service {
	return &service{repo: repo}
}
