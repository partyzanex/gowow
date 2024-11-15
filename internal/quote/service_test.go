package quote_test

import (
	"context"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/partyzanex/gowow/internal/quote"
)

func TestServiceSuite(t *testing.T) {
	suite.Run(t, &ServiceSuite{})
}

type ServiceSuite struct {
	suite.Suite

	ctrl       *gomock.Controller
	repository *MockRepository
	randomizer *MockRandomizer
	service    *quote.Service
}

func (s *ServiceSuite) BeforeTest(_, _ string) {
	s.ctrl = gomock.NewController(s.T())
	s.repository = NewMockRepository(s.ctrl)
	s.randomizer = NewMockRandomizer(s.ctrl)
	s.service = quote.NewService(s.repository, s.randomizer)
}

func (s *ServiceSuite) AfterTest(_, _ string) {
	s.ctrl.Finish()
}

func (s *ServiceSuite) TestGetRandom() {
	const samples = 1000

	ctx := context.Background()
	want := &quote.Quote{
		Content: "test content",
		Author:  "test author",
	}

	s.repository.EXPECT().
		Count(gomock.Any()).
		Return(samples, nil).
		Times(samples)
	s.randomizer.EXPECT().
		GetRandomNumber(0, samples).
		DoAndReturn(func(start, end int) (int, error) {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(end)))
			if err != nil {
				return 0, err
			}

			return int(n.Int64()), nil
		}).Times(samples)
	s.repository.EXPECT().
		GetByID(gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(_ context.Context, id int) (*quote.Quote, error) {
				s.True(id >= 0)
				s.True(id < samples)

				res := *want

				return &res, nil
			},
		).
		Times(samples)

	for i := 0; i < samples; i++ { //nolint:intrange
		got, err := s.service.GetRandomQuote(ctx)
		s.NoError(err)
		s.NotNil(got)
		s.Equal(want, got)
	}
}

func (s *ServiceSuite) TestGetRandomErrors() {
	ctx := context.Background()
	wantErr := errors.New("test error")

	s.repository.EXPECT().
		Count(gomock.Any()).
		Return(0, wantErr)

	_, err := s.service.GetRandomQuote(ctx)
	s.Error(err)
	s.ErrorIs(err, wantErr)

	s.repository.EXPECT().
		Count(gomock.Any()).
		Return(1, nil)

	s.randomizer.EXPECT().
		GetRandomNumber(0, 1).
		Return(0, wantErr)

	_, err = s.service.GetRandomQuote(ctx)
	s.Error(err)
	s.ErrorIs(err, wantErr)

	s.repository.EXPECT().
		Count(gomock.Any()).
		Return(2, nil)

	s.randomizer.EXPECT().
		GetRandomNumber(0, 2).
		Return(1, nil)

	s.repository.EXPECT().
		GetByID(gomock.Any(), 1).
		Return(nil, wantErr)

	_, err = s.service.GetRandomQuote(ctx)
	s.Error(err)
	s.ErrorIs(err, wantErr)
}
