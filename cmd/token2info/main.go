package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	gopcsc "github.com/go-ctap/pcsc"
	token2pcsc "github.com/go-ctap/token2/transport/pcsc"
)

func main() {
	ctx := context.Background()
	readerFlag := flag.String("reader", "", "PC/SC reader name")
	flag.Parse()

	reader := *readerFlag
	if reader == "" {
		var readers []string
		for info, err := range gopcsc.Enumerate() {
			check(err)

			fmt.Printf("reader: %s\n", info.Name)
			readers = append(readers, info.Name)
		}

		if len(readers) != 1 {
			check(fmt.Errorf("found %d readers; select one with -reader", len(readers)))
		}

		reader = readers[0]
	}

	device, err := token2pcsc.Open(reader)
	check(err)

	defer func() {
		_ = device.Close()
	}()

	atr, err := device.ATRInfo(ctx)
	check(err)

	fmt.Printf("ATR: %x\n", atr.Raw)
	fmt.Printf("serial suffix: %s\n", atr.SerialSuffix)
	fmt.Printf("product ID: %04x\n", atr.ProductID)

	serialNumber, err := device.SerialNumber(ctx)
	if err != nil {
		fmt.Printf("serial number: unsupported or failed: %v\n", err)
	} else {
		fmt.Printf("serial number: %s\n", serialNumber)
	}

	config, err := device.Config(ctx)
	if err != nil {
		fmt.Printf("configuration: unsupported or failed: %v\n", err)
	} else if len(config.Raw) == 1 {
		fmt.Printf("configuration: legacy transfer type=%02x\n", config.TransferType)
	} else {
		fmt.Printf("appearance: %x\n", config.Appearance)
		fmt.Printf("FIDO version: %d.%d.%d\n", config.FIDOVersion.Major, config.FIDOVersion.Minor, config.FIDOVersion.Patch)
		fmt.Printf("capabilities: hotp=%t totp=%t nfc=%t ccid=%t fingerprint=%t fido2.1=%t\n",
			config.HOTPSupported(), config.TOTPSupported(), config.NFCSupported(), config.CCIDSupported(),
			config.FingerprintSensorPresent(), config.FIDO21Supported())
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "token2info:", err)
		os.Exit(1)
	}
}
