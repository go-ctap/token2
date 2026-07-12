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

	// FIDOInfo is intentionally first: each public operation must work on a
	// freshly opened connection rather than depend on an earlier call.
	fidoInfo, err := device.FIDOInfo()
	require.NoError(t, err)
	require.NotEmpty(t, fidoInfo)

	config, err := device.Config()
	require.NoError(t, err)
	require.NotEmpty(t, config.Raw)

	serialNumber, err := device.SerialNumber()
	require.NoError(t, err)
	require.NotEmpty(t, serialNumber)

	atr, err := device.ATRInfo()
	require.NoError(t, err)
	require.NotEmpty(t, atr.Raw)
}
