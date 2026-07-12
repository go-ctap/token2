package hid

import (
	"os"
	"testing"

	"github.com/go-ctap/token2"
	"github.com/stretchr/testify/require"
)

func TestHardware(t *testing.T) {
	path := os.Getenv("TOKEN2_HID_TEST_PATH")
	if path == "" {
		t.Skip("set TOKEN2_HID_TEST_PATH to run the HID hardware test")
	}

	device, err := Open(path)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, device.Close())
	})

	serialNumber, err := device.SerialNumber()
	require.NoError(t, err)
	require.NotEmpty(t, serialNumber)

	_, identified := token2.Identify(serialNumber)
	require.True(t, identified)
}
