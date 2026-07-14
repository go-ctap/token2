package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	githubhid "github.com/go-ctap/hid"
	"github.com/go-ctap/token2"
	token2hid "github.com/go-ctap/token2/transport/hid"
)

const (
	token2VendorID = 0x349e
	fidoUsagePage  = 0xf1d0
	fidoUsage      = 0x01
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) (err error) {
	device, path, serialNumber, err := openDevice(ctx, os.Getenv("TOKEN2_HID_PATH"))
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, device.Close())
	}()

	fmt.Printf("HID path: %s\n", path)
	fmt.Printf("Serial number: %s\n", serialNumber)
	if identity, ok := token2.Identify(serialNumber); ok {
		fmt.Printf("Model: %s\n", identity.Model.DisplayName())
	}

	return nil
}

func openDevice(ctx context.Context, path string) (*token2hid.Device, string, string, error) {
	if path != "" {
		device, err := token2hid.Open(path)
		if err != nil {
			return nil, "", "", fmt.Errorf("open HID path %q: %w", path, err)
		}

		serialNumber, err := device.SerialNumber(ctx)
		if err != nil {
			return nil, "", "", errors.Join(
				fmt.Errorf("read serial number from HID path %q: %w", path, err),
				device.Close(),
			)
		}
		return device, path, serialNumber, nil
	}

	for info, err := range githubhid.Enumerate(githubhid.WithVendorID(token2VendorID)) {
		if err != nil {
			return nil, "", "", fmt.Errorf("enumerate HID devices: %w", err)
		}
		if err := ctx.Err(); err != nil {
			return nil, "", "", err
		}
		if info.UsagePage == fidoUsagePage && info.Usage == fidoUsage {
			continue
		}

		device, err := token2hid.Open(info.Path)
		if err != nil {
			continue
		}
		serialNumber, serialErr := device.SerialNumber(ctx)
		if serialErr != nil {
			_ = device.Close()
			continue
		}

		return device, info.Path, serialNumber, nil
	}

	return nil, "", "", errors.New("no Token2 feature HID interface found")
}
