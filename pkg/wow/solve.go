package wow

import (
	"context"
	"crypto/sha256"
	"encoding/binary"

	"github.com/partyzanex/gowow/pkg/proto"
)

// solveChallenge solves the Proof of Work challenge.
func solveChallenge(ctx context.Context, prefix []byte, difficulty uint8) ([]byte, error) {
	// 8 bytes for nonce.
	nonce := make([]byte, 8) //nolint:mnd
	data := make([]byte, len(prefix)+len(nonce))
	copy(data, prefix) // Copy prefix to start of data.

	i := uint64(0)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		binary.LittleEndian.PutUint64(nonce, i)
		copy(data[len(prefix):], nonce)

		hash := sha256.Sum256(data)

		if proto.HasLeadingZeroBits(hash[:], difficulty) {
			return nonce, nil
		}

		i++
	}
}
