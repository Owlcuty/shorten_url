package hash

import (
	"github.com/spaolacci/murmur3"
)

type Murmur struct {
}

func NewMurmur() *Murmur {
	return &Murmur{}
}

func (m *Murmur) Generate(src string) (uint64, error) {
	if src == "" {
		return 0, nil
	}

	num := murmur3.Sum64([]byte(src))
	return num, nil
}
