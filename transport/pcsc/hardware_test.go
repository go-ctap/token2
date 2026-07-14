package pcsc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHardware(t *testing.T) {
	reader := os.Getenv("TOKEN2_PCSC_TEST_READER")
	if reader == "" {
		t.Skip("set TOKEN2_PCSC_TEST_READER to run the PC/SC hardware test")
	}

	device, err := Open(reader)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, device.Close())
	})

	// SerialNumber is intentionally first: it must perform its internal
	// configuration query on a fresh connection.
	serialNumber, err := device.SerialNumber(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, serialNumber)

	config, err := device.Config(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, config.Raw)

	atr, err := device.ATRInfo(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, atr.Raw)
}
