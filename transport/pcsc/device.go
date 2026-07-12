// Package pcsc provides access to Token2 devices over PC/SC.
package pcsc

import (
	"sync"

	nativepcsc "github.com/go-ctap/pcsc"
	"github.com/go-ctap/token2"
	"github.com/go-ctap/token2/apdu"
	"github.com/go-ctap/token2/internal/protocol"
)

var (
	_ token2.SerialNumberDevice = (*Device)(nil)
	_ token2.ATRDevice          = (*Device)(nil)
)

// Device is a Token2 device connected through a PC/SC reader.
type Device struct {
	mu   sync.Mutex
	card nativepcsc.Card
}

// Open connects to the Token2 device in reader.
func Open(reader string) (*Device, error) {
	card, err := nativepcsc.Open(reader)
	if err != nil {
		return nil, err
	}

	return &Device{card: card}, nil
}

// Close closes the PC/SC connection.
func (d *Device) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.card.Close()
}

// ATRInfo returns information encoded in the device ATR.
func (d *Device) ATRInfo() (token2.ATRInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	status, err := d.card.Status()
	if err != nil {
		return token2.ATRInfo{}, err
	}

	return token2.ParseATR(status.ATR)
}

// Config returns the Token2 device configuration.
func (d *Device) Config() (token2.Config, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.config()
}

func (d *Device) config() (token2.Config, error) {
	response, err := apdu.Exchange(d.card, protocol.SelectOTPCommand())
	if err != nil {
		return token2.Config{}, err
	}
	if err := response.Err("select Token2 OTP application"); err != nil {
		return token2.Config{}, err
	}

	response, err = apdu.Exchange(d.card, protocol.ConfigCommand())
	if err != nil {
		return token2.Config{}, err
	}
	if err := response.Err("read Token2 configuration"); err != nil {
		return token2.Config{}, err
	}

	return token2.ParseConfig(response.Data)
}

// FIDOInfo returns the raw FIDO information reported by the device.
func (d *Device) FIDOInfo() ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.fidoInfo()
}

func (d *Device) fidoInfo() ([]byte, error) {
	response, err := apdu.Exchange(d.card, protocol.FIDOInfoCommand())
	if err != nil {
		return nil, err
	}
	if err := response.Err("read FIDO information"); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// SerialNumber returns the full device serial number.
func (d *Device) SerialNumber() (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, err := d.config(); err != nil {
		return "", err
	}
	if _, err := d.fidoInfo(); err != nil {
		return "", err
	}

	response, err := apdu.Exchange(d.card, protocol.SerialNumberCommand(false))
	if err != nil {
		return "", err
	}
	if err := response.Err("read serial number"); err != nil {
		return "", err
	}

	return token2.ParseSerialNumber(response.Data)
}
