package rng

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type Randomizer struct{}

func NewRandomizer() *Randomizer {
	return &Randomizer{}
}

// GetRandomBytes generates a random byte slice of the specified length.
func (r *Randomizer) GetRandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, fmt.Errorf("invalid number of bytes: %d", n)
	}

	res := make([]byte, n)

	_, err := rand.Read(res)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return res, nil
}

// GetRandomNumber generates a random number between the start and end values.
func (*Randomizer) GetRandomNumber(start, end int) (int, error) {
	if start >= end {
		return 0, fmt.Errorf("invalid range: start (%d) >= end (%d)", start, end)
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(end-start)))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}

	return int(n.Int64()) + start, nil
}
