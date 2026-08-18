package hash

import (
	"fmt"
	"shorty/internal/domain"
)

type urlHasher struct {
	generator domain.Generator
	encoder   domain.Encoder
}

func NewURLHasher(g domain.Generator, e domain.Encoder) *urlHasher {
	return &urlHasher{generator: g, encoder: e}
}

func (hasher *urlHasher) Hash(src string) (string, error) {
	num, err := hasher.generator.Generate(src)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash: %w", err)
	}

	hash, err := hasher.encoder.Encode(num)
	if err != nil {
		return "", fmt.Errorf("failed to encode hash: %s", err)
	}

	return hash, err
}
