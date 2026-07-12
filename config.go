package token2

import (
	"errors"
	"fmt"
)

const (
	deviceConfigHOTPSupported            = 0x04
	deviceConfigFingerprintSensorPresent = 0x08
	deviceConfigNFCSupported             = 0x10

	deviceExtensionTOTPSupported   = 0x01
	deviceExtensionFIDO21Supported = 0x02
	deviceExtensionCCIDSupported   = 0x10
)

// ErrInvalidConfigResponse reports malformed configuration data received from
// a Token2 device.
var ErrInvalidConfigResponse = errors.New("token2: invalid configuration response")

// Version is a three-component firmware or protocol version.
type Version struct {
	Major byte
	Minor byte
	Patch byte
}

// Config describes the Token2 configuration response. Raw retains the complete
// response, including fields unknown to this version of the package.
type Config struct {
	Raw []byte

	TransferType        byte
	DeviceConfiguration byte
	Appearance          [4]byte
	FIDOVersion         Version
	DeviceExtension     byte
}

// HOTPSupported reports whether the device supports HOTP.
func (c Config) HOTPSupported() bool {
	return c.DeviceConfiguration&deviceConfigHOTPSupported != 0
}

// FingerprintSensorPresent reports whether the device has a fingerprint sensor.
func (c Config) FingerprintSensorPresent() bool {
	return c.DeviceConfiguration&deviceConfigFingerprintSensorPresent != 0
}

// NFCSupported reports whether the device supports NFC.
func (c Config) NFCSupported() bool {
	return c.DeviceConfiguration&deviceConfigNFCSupported != 0
}

// TOTPSupported reports whether the device supports TOTP.
func (c Config) TOTPSupported() bool {
	return c.DeviceExtension&deviceExtensionTOTPSupported != 0
}

// FIDO21Supported reports whether the device supports FIDO 2.1.
func (c Config) FIDO21Supported() bool {
	return c.DeviceExtension&deviceExtensionFIDO21Supported != 0
}

// CCIDSupported reports whether the device exposes a CCID interface.
func (c Config) CCIDSupported() bool {
	return c.DeviceExtension&deviceExtensionCCIDSupported != 0
}

// ParseConfig parses either a one-byte legacy response or a modern response of
// at least ten bytes. Additional bytes are retained in Raw.
func ParseConfig(response []byte) (Config, error) {
	if len(response) == 0 || len(response) > 1 && len(response) < 10 {
		return Config{}, fmt.Errorf(
			"%w: got %d bytes, want 1 or at least 10",
			ErrInvalidConfigResponse,
			len(response),
		)
	}

	config := Config{
		Raw:          response,
		TransferType: response[0],
	}
	if len(response) == 1 {
		return config, nil
	}

	config.DeviceConfiguration = response[1]
	config.Appearance = [4]byte(response[2:6])
	config.FIDOVersion = Version{response[6], response[7], response[8]}
	config.DeviceExtension = response[9]

	return config, nil
}
