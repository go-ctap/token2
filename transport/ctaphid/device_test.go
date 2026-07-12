package ctaphid

import (
	"bytes"
	"errors"
	"io"
	"testing"

	lowlevel "github.com/go-ctap/ctap/transport/ctaphid"
	"github.com/go-ctap/token2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	reportSize    = 65
	initPacketBit = 0x80
)

type scriptedDevice struct {
	reads  *bytes.Reader
	writes bytes.Buffer
	closed bool
}

func (d *scriptedDevice) Read(p []byte) (int, error) {
	return d.reads.Read(p)
}

func (d *scriptedDevice) Write(p []byte) (int, error) {
	return d.writes.Write(p)
}

func (d *scriptedDevice) Close() error {
	d.closed = true
	return nil
}

func TestATRInfo(t *testing.T) {
	nonce := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	cid := lowlevel.ChannelID{1, 2, 3, 4}
	atr := []byte{
		0x3b, 0xff, 0x18, 0x00, 0x00, 0x10, 0x80,
		0x86, 0x8e, 0x00, 0x16, 0x60, 0x00, 0x60,
		'3', '5', '7', '8', '0', '5', '2', '8',
	}

	initData := append([]byte(nil), nonce...)
	initData = append(initData, cid[:]...)
	initData = append(initData, 2, 1, 2, 3, 0)
	device := newScriptedDevice(t,
		responseBytes(t, lowlevel.BROADCAST_CID, lowlevel.CTAPHID_INIT, initData),
		responseBytes(t, cid, CommandGetATR, atr),
	)

	transport, err := initialize(device, bytes.NewReader(nonce))
	require.NoError(t, err)

	info, err := transport.ATRInfo()
	require.NoError(t, err)
	assert.Equal(t, uint16(0x0016), info.ProductID)
	assert.Equal(t, "35780528", info.SerialSuffix)
	assert.Equal(t, atr, info.Raw)

	written := device.writes.Bytes()
	require.Len(t, written, 2*reportSize)
	assert.Equal(t, byte(lowlevel.CTAPHID_INIT)|initPacketBit, written[5])
	assert.Equal(t, nonce, written[8:16])
	assert.Equal(t, byte(CommandGetATR)|initPacketBit, written[reportSize+5])
	assert.Equal(t, cid[:], written[reportSize+1:reportSize+5])
}

func TestATRInfoRejectsMalformedResponse(t *testing.T) {
	cid := lowlevel.ChannelID{1, 2, 3, 4}
	device := newScriptedDevice(t, responseBytes(t, cid, CommandGetATR, []byte{1, 2, 3}))
	transport := &Device{device: device, cid: cid}

	_, err := transport.ATRInfo()

	assert.ErrorIs(t, err, token2.ErrInvalidATR)
}

func TestInitializeClosesDeviceOnFailure(t *testing.T) {
	device := &scriptedDevice{reads: bytes.NewReader(nil)}

	_, err := initialize(device, bytes.NewReader(make([]byte, 8)))

	require.Error(t, err)
	assert.True(t, device.closed)
}

func TestInitializeReportsRandomFailure(t *testing.T) {
	device := &scriptedDevice{reads: bytes.NewReader(nil)}
	randomErr := errors.New("random failed")

	_, err := initialize(device, errorReader{err: randomErr})

	assert.ErrorIs(t, err, randomErr)
	assert.True(t, device.closed)
}

func TestClose(t *testing.T) {
	device := &scriptedDevice{reads: bytes.NewReader(nil)}
	transport := &Device{device: device}

	require.NoError(t, transport.Close())
	assert.True(t, device.closed)
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

func newScriptedDevice(t testing.TB, responses ...[]byte) *scriptedDevice {
	t.Helper()
	return &scriptedDevice{reads: bytes.NewReader(bytes.Join(responses, nil))}
}

func responseBytes(t testing.TB, cid lowlevel.ChannelID, command lowlevel.Command, data []byte) []byte {
	t.Helper()

	message, err := lowlevel.NewMessage(cid, command, data)
	require.NoError(t, err)

	var reports bytes.Buffer
	_, err = message.WriteTo(&reports)
	require.NoError(t, err)

	encoded := reports.Bytes()
	var response []byte
	for len(encoded) > 0 {
		require.GreaterOrEqual(t, len(encoded), reportSize)
		response = append(response, encoded[1:reportSize]...)
		encoded = encoded[reportSize:]
	}

	return response
}

var _ io.ReadWriteCloser = (*scriptedDevice)(nil)
