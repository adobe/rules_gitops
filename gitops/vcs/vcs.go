package vcs

type VCS interface {
	CreatePR(from, to, title string) error
}

type VCSFunc func(from, to, title string) error

func (f VCSFunc) CreatePR(from, to, title string) error {
	return f(from, to, title)
}
