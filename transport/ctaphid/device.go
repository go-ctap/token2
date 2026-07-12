// Package ctaphid accesses Token2 vendor commands over the CTAPHID protocol.
package ctaphid

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"sync"

	lowlevel "github.com/go-ctap/ctap/transport/ctaphid"
	hidapi "github.com/go-ctap/hid"
	"github.com/go-ctap/token2"
)

// CommandGetATR is the Token2 vendor command which returns the device ATR.
// Its logical CTAPHID value is 0x41; CTAPHID framing adds the init-packet bit,
// resulting in the on-wire command byte 0xc1.
const CommandGetATR lowlevel.Command = lowlevel.CTAPHID_VENDOR_FIRST + 1

// Device is a Token2 device connected through the CTAPHID protocol.
type Device struct {
	mu     sync.Mutex
	device io.ReadWriteCloser
	cid    lowlevel.ChannelID
}

var _ token2.ATRDevice = (*Device)(nil)

// Open opens the FIDO HID collection at path and allocates a CTAPHID channel.
func Open(path string) (*Device, error) {
	device, err := hidapi.OpenPath(path)
	if err != nil {
		return nil, err
	}

	return initialize(device, rand.Reader)
}

func initialize(device io.ReadWriteCloser, random io.Reader) (*Device, error) {
	nonce := make([]byte, 8)
	if _, err := io.ReadFull(random, nonce); err != nil {
		return nil, closeOnError(device, fmt.Errorf("generate CTAPHID nonce: %w", err))
	}

	response, err := lowlevel.Init(device, lowlevel.BROADCAST_CID, nonce)
	if err != nil {
		return nil, closeOnError(device, fmt.Errorf("initialize CTAPHID channel: %w", err))
	}

	return &Device{device: device, cid: response.CID}, nil
}

func closeOnError(device io.Closer, err error) error {
	return errors.Join(err, device.Close())
}

// ATRInfo returns the ATR supplied by Token2 CTAPHID vendor command 0x41.
func (d *Device) ATRInfo() (token2.ATRInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	response, err := lowlevel.Vendor(d.device, d.cid, CommandGetATR, nil)
	if err != nil {
		return token2.ATRInfo{}, fmt.Errorf("read Token2 ATR over CTAPHID: %w", err)
	}

	return token2.ParseATR(response.Data)
}

// Close closes the underlying HID device.
func (d *Device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.device.Close()
}
