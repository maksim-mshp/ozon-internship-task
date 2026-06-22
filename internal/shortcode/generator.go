package shortcode

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var (
	ErrEmptyAlphabet = errors.New("shortcode: alphabet must not be empty")
	ErrInvalidLength = errors.New("shortcode: length must be positive")
)

type Generator struct {
	alphabet []byte
	length   int
}

func NewGenerator(alphabet string, length int) (*Generator, error) {
	if length <= 0 {
		return nil, ErrInvalidLength
	}
	if alphabet == "" {
		return nil, ErrEmptyAlphabet
	}
	return &Generator{
		alphabet: []byte(alphabet),
		length:   length,
	}, nil
}

func (g *Generator) Generate() (string, error) {
	bound := big.NewInt(int64(len(g.alphabet)))
	code := make([]byte, g.length)
	for i := range code {
		idx, err := rand.Int(rand.Reader, bound)
		if err != nil {
			return "", err
		}
		code[i] = g.alphabet[idx.Int64()]
	}
	return string(code), nil
}

func (g *Generator) Length() int {
	return g.length
}
