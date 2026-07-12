package token2

import (
	"errors"
	"fmt"
)

const serialSuffixLength = 8

var historicalPrefix = [2]byte{0x86, 0x8e}

// ErrInvalidATR reports an ATR that does not contain the expected Token2
// historical bytes.
var ErrInvalidATR = errors.New("token2: invalid ATR")

// ATRInfo contains Token2 identity data encoded in a PC/SC ATR.
type ATRInfo struct {
	Raw          []byte
	ProductID    uint16
	SerialSuffix string
}

// ParseATR extracts the USB product ID and decimal serial suffix from a Token2
// ATR. Raw refers to the supplied ATR slice.
func ParseATR(atr []byte) (ATRInfo, error) {
	if len(atr) < 22 {
		return ATRInfo{}, fmt.Errorf("%w: got %d bytes, need at least 22", ErrInvalidATR, len(atr))
	}

	historical := atr[7:22]
	if historical[0] != historicalPrefix[0] || historical[1] != historicalPrefix[1] {
		return ATRInfo{}, fmt.Errorf("%w: unexpected historical prefix %x", ErrInvalidATR, historical[:2])
	}

	serialOffset := len(historical) - serialSuffixLength

	return ATRInfo{
		Raw:          atr,
		ProductID:    uint16(historical[2])<<8 | uint16(historical[3]),
		SerialSuffix: string(historical[serialOffset:]),
	}, nil
}
