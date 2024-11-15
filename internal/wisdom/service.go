package wisdom

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/partyzanex/gowow/pkg/proto"
)

const (
	minDifficulty = 1
	maxDifficulty = 32
)

//go:generate mockgen --source=./service.go -destination=./service_mocks_test.go -package=wisdom_test

type Randomizer interface {
	GetRandomBytes(n int) ([]byte, error)
}

type QuoteService interface {
	GetRandomQuote(ctx context.Context) (*proto.Quote, error)
}

// Service provides functionality for generating challenges and validating solutions.
type Service struct {
	randomizer   Randomizer
	quoteService QuoteService

	timeout     time.Duration
	difficulty  uint8
	randomBytes int
}

// NewService creates a new instance of the Service with the provided parameters.
func NewService(
	randomizer Randomizer,
	quoteService QuoteService,
	timeout time.Duration,
	difficulty, randomBytes uint,
) (*Service, error) {
	if difficulty > maxDifficulty || difficulty < minDifficulty {
		return nil, fmt.Errorf("difficulty must be between %d and %d", minDifficulty, maxDifficulty)
	}

	return &Service{
		randomizer:   randomizer,
		quoteService: quoteService,
		timeout:      timeout,
		difficulty:   uint8(difficulty),
		randomBytes:  int(randomBytes), //nolint:gosec
	}, nil
}

// GenerateChallenge generates a new challenge.
func (s *Service) GenerateChallenge(_ context.Context) (task *proto.Task, deadline *time.Time, err error) {
	prefix, err := s.randomizer.GetRandomBytes(s.randomBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get random bytes: %w", err)
	}

	datetime := time.Now().Add(s.timeout)

	return &proto.Task{
		Prefix:     prefix,
		Difficulty: s.difficulty,
	}, &datetime, nil
}

// GetWisdom validates the solution and returns a quote.
func (s *Service) GetWisdom(ctx context.Context, task *proto.Task, solution *proto.Solution) (*proto.Quote, error) {
	if task == nil {
		return nil, errors.New("task is nil")
	}

	if solution == nil {
		return nil, errors.New("solution is nil")
	}

	if len(task.Prefix) != s.randomBytes {
		return nil, fmt.Errorf("invalid task prefix length: expected %d, got %d", s.randomBytes, len(task.Prefix))
	}

	// Check nonce length.
	if len(solution.Nonce) == 0 {
		return nil, errors.New("solution nonce is empty")
	}

	hash := sha256.Sum256(append(task.Prefix, solution.Nonce...))

	// PoW target is the first difficulty bits of the hash.
	if !proto.HasLeadingZeroBits(hash[:], s.difficulty) {
		return nil, errors.New("invalid solution")
	}

	quote, err := s.quoteService.GetRandomQuote(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get random quote: %w", err)
	}

	return quote, nil
}
