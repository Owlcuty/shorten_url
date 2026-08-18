package domain

type Tools struct {
	DB     URLRepository
	Cache  URLRepository
	Hasher Hasher
}
