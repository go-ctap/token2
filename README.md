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
| Raw FIDO information query | Yes | No | No |

The PC/SC serial-number query performs the device-specific configuration and
FIDO information prelude required by supported Token2 devices. Some proprietary
queries are not available on every Token2 generation; `ATRInfo` remains the
portable PC/SC identity source.

## PC/SC

```go
import token2pcsc "github.com/go-ctap/token2/transport/pcsc"

device, err := token2pcsc.Open("TOKEN2 FIDO2 Security Key(0016)")
if err != nil {
	log.Fatal(err)
}
defer device.Close()

info, err := device.ATRInfo()
if err != nil {
	log.Fatal(err)
}
log.Printf("pid=%04x serial suffix=%s", info.ProductID, info.SerialSuffix)
```

## HID

```go
import token2hid "github.com/go-ctap/token2/transport/hid"

device, err := token2hid.Open(path)
if err != nil {
	log.Fatal(err)
}
defer device.Close()

serialNumber, err := device.SerialNumber()
if err != nil {
	log.Fatal(err)
}
```

## CTAPHID

The CTAPHID transport sends logical vendor command `0x41`; CTAPHID framing adds
the init-packet bit, so the on-wire command byte is `0xc1`.

```go
import token2ctaphid "github.com/go-ctap/token2/transport/ctaphid"

device, err := token2ctaphid.Open(path)
if err != nil {
	log.Fatal(err)
}
defer device.Close()

info, err := device.ATRInfo()
if err != nil {
	log.Fatal(err)
}
```

All concrete device types serialize complete logical operations. Malformed data
received from a card or HID device is returned as an error. Callers are expected
to pass valid reader names, HID paths, serial-number strings and APDU commands;
the package does not add defensive checks for programmer misuse.

## Hardware tests

Hardware tests are opt-in:

```sh
TOKEN2_PCSC_TEST_READER='TOKEN2 FIDO2 Security Key(0016)' go test -run TestHardware -v ./transport/pcsc
TOKEN2_HID_TEST_PATH='platform-specific-path' go test -run TestHardware -v ./transport/hid
TOKEN2_CTAPHID_TEST_PATH='platform-specific-path' go test -run TestHardware -v ./transport/ctaphid
```
