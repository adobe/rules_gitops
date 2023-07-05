package git

type Server interface {
	CreatePR(from, to, title, body string) error
}

type ServerFunc func(from, to, title, body string) error

func (f ServerFunc) CreatePR(from, to, title, body string) error {
	if body == "" {
		body = title
	}

	return f(from, to, title, body)
}
