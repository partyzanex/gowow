package tcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/partyzanex/gowow/pkg/proto"
)

//go:generate mockgen --source=./service.go -destination=./service_mocks_test.go -package=tcp_test

type WisdomService interface {
	GenerateChallenge(ctx context.Context) (task *proto.Task, deadline *time.Time, err error)
	GetWisdom(ctx context.Context, task *proto.Task, solution *proto.Solution) (*proto.Quote, error)
}

type Logger interface {
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
}

type Service struct {
	wisdomService WisdomService

	wg           sync.WaitGroup
	mu           sync.Mutex
	shutdownChan chan struct{}

	listener net.Listener
	logger   Logger
}

func NewService(
	wisdomService WisdomService,
	logger Logger,
	addr string,
) (*Service, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen: %w", err)
	}

	return &Service{
		wisdomService: wisdomService,
		wg:            sync.WaitGroup{},
		mu:            sync.Mutex{},
		shutdownChan:  make(chan struct{}),
		listener:      listener,
		logger:        logger,
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-s.shutdownChan:
			return nil
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			s.logger.Error("cannot accept connection", "error", err)
			continue
		}

		s.logger.Debug("accepted connection")

		go s.handleConn(conn) //nolint:contextcheck
	}
}

func (s *Service) handleConn(conn net.Conn) {
	ts := time.Now()
	defer func() {
		s.logger.Debug("request duration", "duration", time.Since(ts))
	}()

	select {
	case <-s.shutdownChan:
		return
	default:
	}

	release := s.getReleaseConnFn(conn)
	defer release()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.watchContext(ctx, cancel)

	if err := s.challenge(ctx, conn); err != nil {
		s.logger.Error("challenge failed", "error", err)
		return
	}
}

func (s *Service) challenge(ctx context.Context, conn net.Conn) error {
	encoder := json.NewEncoder(conn)

	task, deadline, err := s.wisdomService.GenerateChallenge(ctx)
	if err != nil {
		return s.handleErr(fmt.Errorf("cannot generate challenge: %w", err), encoder)
	}

	if err = conn.SetDeadline(*deadline); err != nil {
		return fmt.Errorf("cannot set deadline: %w", err)
	}

	s.logger.Debug("sending challenge task", "task", task)

	if err = encoder.Encode(task); err != nil {
		return fmt.Errorf("cannot encode task: %w", err)
	}

	decoder := json.NewDecoder(conn)
	solution := new(proto.Solution)

	if err = decoder.Decode(solution); err != nil {
		return s.handleErr(fmt.Errorf("cannot decode solution: %w", err), encoder)
	}

	s.logger.Debug("received solution", "solution", solution)

	ctx, cancel := context.WithDeadline(ctx, *deadline)
	defer cancel()

	quote, err := s.wisdomService.GetWisdom(ctx, task, solution)
	if err != nil {
		return s.handleErr(fmt.Errorf("cannot get wisdom: %w", err), encoder)
	}

	s.logger.Debug("sending quote", "quote", quote)

	if err = encoder.Encode(&proto.Result{Quote: quote}); err != nil {
		return fmt.Errorf("cannot encode quote: %w", err)
	}

	return nil
}

func (s *Service) handleErr(err error, encoder *json.Encoder) error {
	if err == nil {
		s.logger.Warn("error is nil")
		return nil
	}

	errMsg := err.Error()

	if encodeErr := encoder.Encode(&proto.Result{Error: &errMsg}); encodeErr != nil {
		return errors.Join(err, fmt.Errorf("cannot encode error: %w", encodeErr))
	}

	return err
}

func (s *Service) take() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.shutdownChan:
		return
	default:
	}

	s.wg.Add(1)
}

func (s *Service) release() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.shutdownChan:
		return
	default:
	}

	s.wg.Done()
}

func (s *Service) Close() error {
	select {
	case <-s.shutdownChan:
		return nil
	default:
	}

	close(s.shutdownChan)

	s.mu.Lock()
	s.wg.Wait()
	s.mu.Unlock()

	return s.listener.Close() //nolint:wrapcheck
}

func (s *Service) watchContext(ctx context.Context, cancel context.CancelFunc) {
	select {
	case <-s.shutdownChan:
		cancel()
	case <-ctx.Done():
	}
}

func (s *Service) getReleaseConnFn(conn net.Conn) func() {
	var once sync.Once

	return func() {
		once.Do(func() {
			s.take()
			defer s.release()

			if err := conn.Close(); err != nil {
				s.logger.Error("cannot close connection: %w", err)
			}
		})
	}
}
