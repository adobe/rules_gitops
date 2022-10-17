package git

type Server interface {
	CreatePR(from, to, title string) error
}

type ServerFunc func(from, to, title string) error

func (f ServerFunc) CreatePR(from, to, title string) error {
	return f(from, to, title)
}

