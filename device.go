package token2

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-ctap/token2/internal/protocol"
)

// Device is the lifecycle contract implemented by Token2 transport devices.
// Transport-specific capabilities are described by capability interfaces such
// as SerialNumberDevice and ATRDevice.
type Device interface {
	Close() error
}

// SerialNumberDevice is a Device capable of returning the full Token2 serial
// number.
type SerialNumberDevice interface {
	Device
	SerialNumber(context.Context) (string, error)
}

// ATRDevice is a Device capable of returning Token2 ATR information.
type ATRDevice interface {
	Device
	ATRInfo(context.Context) (ATRInfo, error)
}

// ErrInvalidSerialResponse reports a malformed serial-number response received
// from a Token2 device.
var ErrInvalidSerialResponse = errors.New("token2: invalid serial-number response")

// ParseSerialNumber parses the TAG-LENGTH-VALUE response returned by the
// Token2 serial-number command.
func ParseSerialNumber(response []byte) (string, error) {
	if len(response) < 2 {
		return "", fmt.Errorf("%w: got %d bytes, need at least 2", ErrInvalidSerialResponse, len(response))
	}
	if response[0] != protocol.SerialResponseTag {
		return "", fmt.Errorf("%w: unexpected tag %02x", ErrInvalidSerialResponse, response[0])
	}

	length := int(response[1])
	if length > len(response)-2 {
		return "", fmt.Errorf(
			"%w: declared length %d exceeds payload length %d",
			ErrInvalidSerialResponse,
			length,
			len(response)-2,
		)
	}

	return string(response[2 : 2+length]), nil
}
