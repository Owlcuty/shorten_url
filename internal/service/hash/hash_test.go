package hash

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBase62_Hash(t *testing.T) {
	encoder := NewBase62Hash()

	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"zero value", 0, "a"},
		{"single digit small", 5, "f"},
		{"last alphabet char", 61, "9"},
		{"first two-digit number", 62, "ab"},
		{"middle range number", 1459203, "HLhg"},
		{"large database ID", 56800235584, "aaaaaab"},
		{"max uint64 value", 18446744073709551615, "pIrkgbKrQ8v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := encoder.Encode(tt.input)
			require.NoError(t, err)

			require.Equal(t, tt.expected, actual, "For input %d Expected %s, got %s", tt.input, tt.expected, actual)
		})
	}
}

func TestMurmur3_Hash(t *testing.T) {
	generator := NewMurmur()

	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{"empty string", "", 0},
		{"single character", "a", 9607679276477937801},
		{"classic test", "hello", 14688674573012802306},
		{"longer string", "The quick brown fox jumps over the lazy dog", 16378391709484522348},
		{"cyrillic string", "привет", 4218789751818088067},
		{"numbers and special", "1234567890!@#", 17584055497701432086},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := generator.Generate(tt.input)
			require.NoError(t, err)

			require.Equal(t, tt.expected, actual, "For input %s Expected %d, got %d", tt.input, tt.expected, actual)
		})
	}
}

func TestHash(t *testing.T) {
	generator := NewMurmur()
	encoder := NewBase62Hash()

	hasher := NewURLHasher(generator, encoder)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "a"},
		{"single character", "a", "RDKXBD5qTBl"},
		{"classic test", "hello", "erwBcaRreFr"},
		{"longer string", "The quick brown fox jumps over the lazy dog", "6lqZIU2m3Ft"},
		{"cyrillic string", "привет", "JwX93ytgObf"},
		{"numbers and special", "1234567890!@#", "kyHni17j76u"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := hasher.Hash(tt.input)
			require.NoError(t, err)

			require.Equal(t, tt.expected, actual, "For input %s Expected %d, got %d", tt.input, tt.expected, actual)
		})
	}
}
