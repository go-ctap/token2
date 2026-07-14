# go-token2

Pure-Go Token2 device support over PC/SC, USB HID feature reports and CTAPHID.

> [!WARNING]
> This module is under active development. Its public API may change during `v0.x`.

## Packages

- `token2` contains transport-independent device models, identity lookup,
  response types, parsers and capability interfaces.
- `token2/apdu` implements the APDU subset used by Token2, including automatic
  `GET RESPONSE` chaining.
- `token2/transport/pcsc` opens Token2 devices through
  [`github.com/go-ctap/pcsc`](https://github.com/go-ctap/pcsc).
- `token2/transport/hid` opens the Token2 HID interface through
  [`github.com/go-ctap/hid`](https://github.com/go-ctap/hid).
- `token2/transport/ctaphid` reads the Token2 ATR through CTAPHID vendor
  command `0x41`, using [`github.com/go-ctap/ctap`](https://github.com/go-ctap/ctap).

## Transport capabilities

| Capability | PC/SC | Feature HID | CTAPHID |
| --- | --- | --- | --- |
| Full serial number | Yes | Yes | No |
| Model identification from the serial number | Yes | Yes | No |
| ATR, product ID and serial suffix | Yes | No | Yes |
| Token2 configuration | Yes | No | No |

The PC/SC serial-number query performs the device-specific configuration query
required by supported Token2 devices. On firmware such as R3.1 it also retries
the serial-number command after an internal compatibility prelude. Some
proprietary queries are not available on every Token2 generation; `ATRInfo`
remains the portable PC/SC identity source.

## PC/SC

```go
import (
	"log"

	"github.com/go-ctap/token2"
	token2pcsc "github.com/go-ctap/token2/transport/pcsc"
)

device, err := token2pcsc.Open("TOKEN2 FIDO2 Security Key(0016)")
if err != nil {
	log.Fatal(err)
}
defer device.Close()

atr, err := device.ATRInfo()
if err != nil {
	log.Fatal(err)
}
log.Printf("ATR=%x", atr.Raw)
log.Printf("pid=%04x serial suffix=%s", atr.ProductID, atr.SerialSuffix)

serialNumber, err := device.SerialNumber()
if err != nil {
	log.Fatal(err)
}
log.Printf("serial number=%s", serialNumber)

if identity, ok := token2.Identify(serialNumber); ok {
	log.Printf(
		"model: revision=%s form-factor=%q branding=%q",
		identity.Model.Revision,
		identity.Model.FormFactor,
		identity.Model.Branding,
	)
}

config, err := device.Config()
if err != nil {
	log.Fatal(err)
}
if len(config.Raw) == 1 {
	log.Printf("legacy transfer type=%02x", config.TransferType)
} else {
	log.Printf(
		"appearance=%x FIDO=%d.%d.%d hotp=%t totp=%t nfc=%t ccid=%t fingerprint=%t fido2.1=%t",
		config.Appearance,
		config.FIDOVersion.Major,
		config.FIDOVersion.Minor,
		config.FIDOVersion.Patch,
		config.HOTPSupported(),
		config.TOTPSupported(),
		config.NFCSupported(),
		config.CCIDSupported(),
		config.FingerprintSensorPresent(),
		config.FIDO21Supported(),
	)
}
```

## HID

```go
import (
	"log"

	"github.com/go-ctap/token2"
	token2hid "github.com/go-ctap/token2/transport/hid"
)

device, err := token2hid.Open(path)
if err != nil {
	log.Fatal(err)
}
defer device.Close()

serialNumber, err := device.SerialNumber()
if err != nil {
	log.Fatal(err)
}
log.Printf("serial number=%s", serialNumber)

if identity, ok := token2.Identify(serialNumber); ok {
	log.Printf(
		"model: revision=%s form-factor=%q branding=%q",
		identity.Model.Revision,
		identity.Model.FormFactor,
		identity.Model.Branding,
	)
}
```

## CTAPHID

The CTAPHID transport sends logical vendor command `0x41`; CTAPHID framing adds
the init-packet bit, so the on-wire command byte is `0xc1`.

```go
import (
	"log"

	token2ctaphid "github.com/go-ctap/token2/transport/ctaphid"
)

device, err := token2ctaphid.Open(path)
if err != nil {
	log.Fatal(err)
}
defer device.Close()

info, err := device.ATRInfo()
if err != nil {
	log.Fatal(err)
}
log.Printf("ATR=%x", info.Raw)
log.Printf("pid=%04x serial suffix=%s", info.ProductID, info.SerialSuffix)
```

All concrete device types serialize complete logical operations. Malformed data
received from a card or HID device is returned as an error. Callers are expected
to pass valid reader names, HID paths, serial-number strings and APDU commands;
the package does not add defensive checks for programmer misuse.

## Examples

Each example is an independent Go module, keeping transport-specific
dependencies out of the root module.

| Example | Purpose | Optional configuration |
| --- | --- | --- |
| [`examples/pcsc`](examples/pcsc) | Read identity and configuration over PC/SC | `PCSC_READER` (reader-name substring) |
| [`examples/hid`](examples/hid) | Read the full serial number over HID feature reports | `TOKEN2_HID_PATH` |
| [`examples/ctaphid`](examples/ctaphid) | Read ATR identity over the CTAPHID vendor command | `TOKEN2_CTAPHID_PATH` |

Run an example from its directory:

```sh
cd examples/pcsc
go run .
```

Without an environment variable, an example selects the first matching device
or reader it finds. Set the corresponding variable when multiple Token2 devices
are connected or when automatic HID selection is not available on the host.

## Hardware tests

Hardware tests are opt-in:

```sh
TOKEN2_PCSC_TEST_READER='TOKEN2 FIDO2 Security Key(0016)' go test -run TestHardware -v ./transport/pcsc
TOKEN2_HID_TEST_PATH='platform-specific-path' go test -run TestHardware -v ./transport/hid
TOKEN2_CTAPHID_TEST_PATH='platform-specific-path' go test -run TestHardware -v ./transport/ctaphid
```
