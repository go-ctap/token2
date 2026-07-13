// Package ctaphid accesses Token2 vendor commands over the CTAPHID protocol.
package ctaphid

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-ctap/ctap/transport/ctaphid"
	hidapi "github.com/go-ctap/hid"
	"github.com/go-ctap/token2"
)

// CommandGetATR is the Token2 vendor command which returns the device ATR.
// Its logical CTAPHID value is 0x41; CTAPHID framing adds the init-packet bit,
// resulting in the on-wire command byte 0xc1.
const CommandGetATR = ctaphid.CTAPHID_VENDOR_FIRST + 1

// Device is a Token2 device connected through the CTAPHID protocol.
type Device struct {
	transport *ctaphid.Transport
}

var _ token2.ATRDevice = (*Device)(nil)

// Open opens the FIDO HID collection at path and allocates a CTAPHID channel.
func Open(ctx context.Context, path string) (*Device, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	device, err := hidapi.OpenPath(path)
	if err != nil {
		return nil, err
	}

	transport, err := ctaphid.Open(ctx, device)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("initialize CTAPHID channel: %w", err), device.Close())
	}

	return &Device{transport: transport}, nil
}

// ATRInfo returns the ATR supplied by Token2 CTAPHID vendor command 0x41.
func (d *Device) ATRInfo(ctx context.Context) (token2.ATRInfo, error) {
	response, err := d.transport.Vendor(ctx, CommandGetATR, nil)
	if err != nil {
		return token2.ATRInfo{}, fmt.Errorf("read Token2 ATR over CTAPHID: %w", err)
	}

	return token2.ParseATR(response.Data)
}

// Close closes the underlying HID device.
func (d *Device) Close() error {
	return d.transport.Close()
}
