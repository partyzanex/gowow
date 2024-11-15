package wow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSolveChallenge(t *testing.T) {
	prefix := []byte("test prefix data")
	difficulty := uint8(16)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nonce, err := solveChallenge(ctx, prefix, difficulty)
	require.NoError(t, err)
	require.NotNil(t, nonce)
	require.Len(t, nonce, 8)

	hash := sha256.Sum256(append(prefix, nonce...))
	require.True(t, bytes.HasPrefix(hash[:2], make([]byte, 2)))
}

func TestSolveChallengeTimeout(t *testing.T) {
	prefix := []byte("test prefix")
	difficulty := uint8(32) // High difficulty to trigger timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	nonce, err := solveChallenge(ctx, prefix, difficulty)
	require.Error(t, err)
	require.Nil(t, nonce)
	require.Equal(t, context.DeadlineExceeded, err)
}

func TestSolveChallengeCancellation(t *testing.T) {
	prefix := []byte("test prefix")
	difficulty := uint8(32)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	nonce, err := solveChallenge(ctx, prefix, difficulty)
	require.Error(t, err)
	require.Nil(t, nonce)
	require.Equal(t, context.Canceled, err)
}

func benchmarkSolveChallenge(b *testing.B, difficulty uint8) {
	b.Helper()
	prefix := []byte("benchmarkPrefix")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _ = solveChallenge(ctx, prefix, difficulty)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := solveChallenge(ctx, prefix, difficulty)
		if err != nil {
			b.Skip()
		}
	}
}

func BenchmarkSolveChallenge(b *testing.B) {
	for i := uint8(1); i <= 20; i++ {
		b.Run(strconv.Itoa(int(i)), func(b *testing.B) {
			benchmarkSolveChallenge(b, i)
		})
	}
}
