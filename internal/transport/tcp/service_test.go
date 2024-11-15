package tcp_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/partyzanex/gowow/internal/transport/tcp"
	"github.com/partyzanex/gowow/pkg/proto"
)

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

type ServiceSuite struct {
	suite.Suite

	ctrl          *gomock.Controller
	wisdomService *MockWisdomService
	logger        *MockLogger
	service       *tcp.Service
	addr          string
}

func (s *ServiceSuite) BeforeTest(_, _ string) {
	s.ctrl = gomock.NewController(s.T())
	s.wisdomService = NewMockWisdomService(s.ctrl)
	s.logger = NewMockLogger(s.ctrl)

	s.logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	s.logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	n, err := rand.Int(rand.Reader, big.NewInt(int64(10000)))
	s.Require().NoError(err)

	s.addr = fmt.Sprintf("localhost:%d", n.Int64())

	tcpService, err := tcp.NewService(s.wisdomService, s.logger, s.addr)
	s.Require().NoError(err)

	s.service = tcpService

	go func() {
		if err := s.service.Run(context.Background()); err != nil {
			panic(err)
		}
	}()
}

func (s *ServiceSuite) AfterTest(_, _ string) {
	s.Require().NoError(s.service.Close())
	s.ctrl.Finish()
}

func (s *ServiceSuite) TestGenerateChallenge() {
	want := &proto.Task{Prefix: []byte("test prefix"), Difficulty: 3}
	deadline := time.Now().Add(10 * time.Second)

	s.wisdomService.EXPECT().
		GenerateChallenge(gomock.Any()).
		Return(want, &deadline, nil)

	conn, err := net.Dial("tcp", s.addr)
	s.Require().NoError(err)

	defer conn.Close()

	decoder := json.NewDecoder(conn)
	got := new(proto.Task)

	err = decoder.Decode(got)
	s.NoError(err)
	s.Equal(want, got)
}

func (s *ServiceSuite) TestGetWisdom() {
	task := &proto.Task{Prefix: []byte("test prefix"), Difficulty: 3}
	deadline := time.Now().Add(10 * time.Second)
	solution := &proto.Solution{Nonce: []byte("valid solution")}
	quote := &proto.Quote{Content: "A wise quote", Author: "Unknown"}

	s.wisdomService.EXPECT().
		GenerateChallenge(gomock.Any()).
		Return(task, &deadline, nil)
	s.wisdomService.EXPECT().
		GetWisdom(gomock.Any(), task, solution).
		Return(quote, nil)

	conn, err := net.Dial("tcp", s.addr)
	s.Require().NoError(err)

	defer conn.Close()

	decoder := json.NewDecoder(conn)
	receivedTask := new(proto.Task)

	err = decoder.Decode(receivedTask)
	s.Require().NoError(err)
	s.Require().Equal(task, receivedTask)

	encoder := json.NewEncoder(conn)

	err = encoder.Encode(solution)
	s.NoError(err)

	result := new(proto.Result)

	err = decoder.Decode(result)
	s.NoError(err)
	s.Equal(quote, result.Quote)
}
