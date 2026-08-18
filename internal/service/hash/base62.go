package hash

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Base62Hash struct {
}

func NewBase62Hash() *Base62Hash {
	return &Base62Hash{}
}

func (b *Base62Hash) Encode(num uint64) (string, error) {
	if num == 0 {
		return string(alphabet[0]), nil
	}

	buf := make([]byte, 0, 12)

	for num > 0 {
		remainder := num % 62

		buf = append(buf, alphabet[remainder])

		num = num / 62
	}

	return string(buf), nil
}
