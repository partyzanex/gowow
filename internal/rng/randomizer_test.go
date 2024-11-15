package rng_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/partyzanex/gowow/internal/rng"
)

func TestRandomizerSuite(t *testing.T) {
	suite.Run(t, new(RandomizerSuite))
}

type RandomizerSuite struct {
	suite.Suite
	randomizer *rng.Randomizer
}

func (s *RandomizerSuite) SetupTest() {
	s.randomizer = rng.NewRandomizer()
}

func (s *RandomizerSuite) TestGetRandomBytes() {
	s.Run("valid number of bytes", func() {
		bytes, err := s.randomizer.GetRandomBytes(10)
		s.NoError(err)
		s.Len(bytes, 10)
	})

	s.Run("invalid number of bytes", func() {
		_, err := s.randomizer.GetRandomBytes(0)
		s.Error(err)
		s.EqualError(err, "invalid number of bytes: 0")
	})
}

func (s *RandomizerSuite) TestGetRandomNumber() {
	s.Run("valid range", func() {
		start, end := 10, 100
		n, err := s.randomizer.GetRandomNumber(start, end)
		s.NoError(err)
		s.GreaterOrEqual(n, start)
		s.Less(n, end)
	})

	s.Run("invalid range", func() {
		_, err := s.randomizer.GetRandomNumber(100, 10)
		s.Error(err)
	})

	s.Run("maximum uint8 range", func() {
		start, end := 0, math.MaxUint8
		n, err := s.randomizer.GetRandomNumber(start, end)
		s.NoError(err)
		s.GreaterOrEqual(n, start)
		s.Less(n, end)
	})
}
