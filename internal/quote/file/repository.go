package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/partyzanex/gowow/internal/quote"
)

//nolint:gochecknoglobals
var (
	sep = []byte(" - ") // quote separator.
	eol = []byte("\n")  // line separator.
)

// Repository represents a repository that holds a collection of quotes loaded from a file.
type Repository struct {
	quotes []*quote.Quote
}

// NewRepository creates a new Repository instance by reading and parsing quotes from the specified file path.
func NewRepository(filePath string) (*Repository, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %q: %w", filePath, err)
	}

	lines := bytes.Split(b, eol)
	count := len(lines)
	quotes := make([]*quote.Quote, count)

	for i, line := range lines {
		quotes[i] = parseQuote(line)
	}

	return &Repository{
		quotes: quotes,
	}, nil
}

func parseQuote(b []byte) *quote.Quote {
	parts := bytes.Split(b, sep)

	if n := len(parts); n >= 2 { //nolint:mnd
		return &quote.Quote{
			Content: string(bytes.Join(parts[:n-1], sep)),
			Author:  string(parts[n-1]),
		}
	}

	return &quote.Quote{
		Content: string(b),
	}
}

// GetByID retrieves a quote by its ID from the repository.
func (r *Repository) GetByID(_ context.Context, id int) (*quote.Quote, error) {
	if id < 0 {
		return nil, errors.New("invalid id")
	}

	if id > len(r.quotes)-1 {
		return nil, errors.New("quote not found")
	}

	return r.quotes[id], nil
}

// Count returns the total number of quotes available in the repository.
func (r *Repository) Count(_ context.Context) (int, error) {
	return len(r.quotes), nil
}
