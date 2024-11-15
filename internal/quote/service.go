package quote

import (
	"context"
	"fmt"

	"github.com/partyzanex/gowow/pkg/proto"
)

//go:generate mockgen --source=./service.go -destination=./service_mocks_test.go -package=quote_test

// Repository defines the interface for a quote repository that provides access to quotes.
type Repository interface {
	GetByID(ctx context.Context, id int) (*Quote, error)
	Count(ctx context.Context) (int, error)
}

// Randomizer defines an interface for generating random numbers within a range.
type Randomizer interface {
	// GetRandomNumber generates a random number between the start and end values.
	GetRandomNumber(start, end int) (int, error)
}

// Service provides functionality for retrieving random quotes.
type Service struct {
	repository Repository
	randomizer Randomizer
}

// Quote is an alias for the Quote type from the proto package, representing a single quote entity.
type Quote = proto.Quote

// NewService creates a new instance of the Service with the provided repository and randomizer.
func NewService(repository Repository, randomizer Randomizer) *Service {
	return &Service{
		repository: repository,
		randomizer: randomizer,
	}
}

// GetRandomQuote retrieves a random quote from the repository.
func (s *Service) GetRandomQuote(ctx context.Context) (*Quote, error) {
	count, err := s.repository.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get quote count: %w", err)
	}

	id, err := s.randomizer.GetRandomNumber(0, count)
	if err != nil {
		return nil, fmt.Errorf("cannot get random quote id: %w", err)
	}

	quote, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cannot get random quote: %w", err)
	}

	return quote, nil
}
