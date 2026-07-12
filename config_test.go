package token2

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	config, err := ParseConfig([]byte{0x02, 0x2a, 0x86, 0x01, 0x10, 0x00, 0x02, 0x01, 0x02, 0x37})
	require.NoError(t, err)
	assert.Equal(t, [4]byte{0x86, 0x01, 0x10, 0x00}, config.Appearance)
	assert.Equal(t, Version{2, 1, 2}, config.FIDOVersion)
	assert.False(t, config.HOTPSupported())
	assert.True(t, config.TOTPSupported())
	assert.True(t, config.CCIDSupported())
}

func TestParseConfigLegacy(t *testing.T) {
	config, err := ParseConfig([]byte{0x02})
	require.NoError(t, err)
	assert.Equal(t, byte(0x02), config.TransferType)
}

func TestParseConfigRejectsTruncatedResponses(t *testing.T) {
	for length := 0; length < 10; length++ {
		if length == 1 {
			continue
		}

		_, err := ParseConfig(make([]byte, length))
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidConfigResponse))
	}
}

func TestParseSerialResponse(t *testing.T) {
	serial, err := ParseSerialNumber([]byte{0xd1, 0x0e, '7', '2', '1', '0', '2', '9', '3', '5', '7', '8', '0', '5', '2', '8'})
	require.NoError(t, err)
	assert.Equal(t, "72102935780528", serial)
}

func TestParseSerialResponseRejectsMalformedData(t *testing.T) {
	tests := [][]byte{
		nil,
		{0xd1},
		{0xd2, 0},
		{0xd1, 2, '1'},
	}

	for _, response := range tests {
		_, err := ParseSerialNumber(response)
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidSerialResponse))
	}
}
