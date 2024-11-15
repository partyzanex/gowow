package wisdom_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/partyzanex/gowow/internal/wisdom"
	"github.com/partyzanex/gowow/pkg/proto"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type WisdomServiceSuite struct {
	suite.Suite

	ctrl          *gomock.Controller
	randomizer    *MockRandomizer
	quoteService  *MockQuoteService
	wisdomService *wisdom.Service
}

func TestWisdomServiceSuite(t *testing.T) {
	suite.Run(t, new(WisdomServiceSuite))
}

func (s *WisdomServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.randomizer = NewMockRandomizer(s.ctrl)
	s.quoteService = NewMockQuoteService(s.ctrl)

	wisdomService, err := wisdom.NewService(s.randomizer, s.quoteService, 10*time.Second, 3, 5)
	s.Require().NoError(err)

	s.wisdomService = wisdomService
}

func (s *WisdomServiceSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *WisdomServiceSuite) TestGenerateChallenge() {
	ctx := context.Background()
	prefix := []byte{0x01, 0x02, 0x03}

	s.randomizer.EXPECT().GetRandomBytes(5).Return(prefix, nil)
	task, deadline, err := s.wisdomService.GenerateChallenge(ctx)

	s.NoError(err)
	s.NotNil(task)
	s.NotNil(deadline)
	s.Equal(prefix, task.Prefix)
	s.Equal(uint8(3), task.Difficulty)
	s.WithinDuration(time.Now().Add(10*time.Second), *deadline, time.Second)
}

func (s *WisdomServiceSuite) TestGetWisdom() {
	ctx := context.Background()

	task := &proto.Task{
		Prefix:     []byte{29, 2, 247, 72, 88},
		Difficulty: 3,
	}

	solution := &proto.Solution{
		Nonce: []byte{198, 134, 229, 0, 0, 0, 0, 0},
	}

	s.quoteService.EXPECT().
		GetRandomQuote(ctx).
		Return(&proto.Quote{Content: "A wise quote", Author: "Author"}, nil)

	quote, err := s.wisdomService.GetWisdom(ctx, task, solution)
	s.NoError(err)
	s.NotNil(quote)
	s.Equal(&proto.Quote{Content: "A wise quote", Author: "Author"}, quote)
}

func (s *WisdomServiceSuite) TestGetWisdomInvalidSolution() {
	ctx := context.Background()
	task := &proto.Task{
		Prefix:     []byte{0x01, 0x03, 0x03, 0x02, 0x03},
		Difficulty: 5,
	}
	solution := &proto.Solution{
		Nonce: []byte{0x07, 0x08, 0x09},
	}

	hash := sha256.New().Sum(solution.Nonce)
	prefix := make([]byte, task.Difficulty)
	if bytes.HasPrefix(prefix, append(task.Prefix, hash...)) {
		s.Fail("solution should be invalid")
	}

	quote, err := s.wisdomService.GetWisdom(ctx, task, solution)
	s.Error(err)
	s.Nil(quote)
	s.Contains(err.Error(), "invalid solution")
}
