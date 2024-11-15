package wow_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/partyzanex/gowow/pkg/proto"
	"github.com/partyzanex/gowow/pkg/wow"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ClientSuite struct {
	suite.Suite

	ctrl    *gomock.Controller
	dialer  *MockDialer
	logger  *MockLogger
	client  *wow.Client
	address string
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

func (s *ClientSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.dialer = NewMockDialer(s.ctrl)
	s.logger = NewMockLogger(s.ctrl)
	s.address = "localhost:8080"

	var err error
	s.client, err = wow.NewClient(
		wow.WithDialer(s.dialer),
		wow.WithLogger(s.logger),
		wow.WithAddress(s.address),
	)
	s.Require().NoError(err)
	s.Require().NotNil(s.client)

	s.logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
}

func (s *ClientSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ClientSuite) TestGetRandomQuote_Success() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn := NewMockConn(s.ctrl)
	task := &proto.Task{Prefix: []byte("prefix"), Difficulty: 2}
	want := &proto.Quote{Content: "A wise quote", Author: "Author"}
	result := &proto.Result{Quote: want}

	s.dialer.EXPECT().
		DialContext(ctx, "tcp", s.address).
		Return(conn, nil)
	conn.EXPECT().Close().Return(nil)

	conn.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(data []byte) (int, error) {
			b, err := json.Marshal(task)
			s.Require().NoError(err)

			return copy(data, b), nil
		})
	conn.EXPECT().Write(gomock.Any()).Return(0, nil)
	conn.EXPECT().
		Read(gomock.Any()).
		DoAndReturn(func(data []byte) (int, error) {
			b, err := json.Marshal(result)
			s.Require().NoError(err)

			return copy(data, b), nil
		})

	got, err := s.client.GetRandomQuote(ctx)
	s.NoError(err)
	s.Equal(want, got)
}

func (s *ClientSuite) TestGetRandomQuote_DialError() {
	ctx := context.Background()
	s.dialer.EXPECT().
		DialContext(ctx, "tcp", s.address).
		Return(nil, errors.New("dial error"))

	result, err := s.client.GetRandomQuote(ctx)
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "cannot dial")
	s.Contains(err.Error(), "dial error")
}

func (s *ClientSuite) TestGetRandomQuote_DecodeTaskError() {
	ctx := context.Background()
	wantErr := errors.New("read error")

	conn := NewMockConn(s.ctrl)
	conn.EXPECT().Close().Return(nil)

	s.dialer.EXPECT().DialContext(ctx, "tcp", s.address).Return(conn, nil)

	conn.EXPECT().
		Read(gomock.Any()).
		Return(0, wantErr)

	result, err := s.client.GetRandomQuote(ctx)
	s.ErrorIs(err, wantErr)
	s.Nil(result)
}
