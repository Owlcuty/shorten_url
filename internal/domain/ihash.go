package domain

type Hasher interface {
	Hash(src string) (string, error)
}

type Generator interface {
	Generate(src string) (uint64, error)
}

type Encoder interface {
	Encode(num uint64) (string, error)
}
