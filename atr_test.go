package token2

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseATR(t *testing.T) {
	tests := []struct {
		atr    []byte
		pid    uint16
		suffix string
	}{
		{[]byte{0x3b, 0xff, 0x18, 0x00, 0x00, 0x10, 0x80, 0x86, 0x8e, 0x00, 0x16, 0x60, 0x00, 0x60, '3', '5', '7', '8', '0', '5', '2', '8'}, 0x0016, "35780528"},
		{[]byte{0x3b, 0xff, 0x18, 0x00, 0x00, 0x81, 0x01, 0x86, 0x8e, 0x02, 0x04, 0x60, 0x00, 0x60, '5', '4', '0', '9', '5', '3', '0', '3', 0x64}, 0x0204, "54095303"},
	}
	for _, tc := range tests {
		info, err := ParseATR(tc.atr)
		require.NoError(t, err)

		assert.Equal(t, tc.pid, info.ProductID)
		assert.Equal(t, tc.suffix, info.SerialSuffix)
	}
}

func TestParseATRRejectsMalformedData(t *testing.T) {
	tests := [][]byte{
		nil,
		make([]byte, 21),
		make([]byte, 22),
	}

	for _, atr := range tests {
		_, err := ParseATR(atr)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidATR))
	}
}

func TestParseATRRejectsGenericPIVATR(t *testing.T) {
	atr := []byte{
		0x3b, 0x8f, 0x80, 0x01, 0x54, 0x4b, 0x00, 0x50, 0x49, 0x56,
		0x04, 0x02, 0x38, 0x38, 0x38, 0x38, 0x38, 0x38, 0x38, 0x60,
	}

	_, err := ParseATR(atr)

	assert.ErrorIs(t, err, ErrInvalidATR)
}
