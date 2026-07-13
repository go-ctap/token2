package main

import (
	"context"
	"fmt"
	"os"

	ghid "github.com/go-ctap/hid"
	"github.com/go-ctap/token2"
	token2hid "github.com/go-ctap/token2/transport/hid"
)

func main() {
	ctx := context.Background()
	for info, err := range ghid.Enumerate(ghid.WithVendorID(0x349e)) {
		check(err)

		fmt.Printf(
			"device: path=%s pid=%04x usage=%04x:%04x product=%q\n",
			info.Path,
			info.ProductID,
			info.UsagePage,
			info.Usage,
			info.ProductStr,
		)

		device, err := token2hid.Open(info.Path)
		if err != nil {
			fmt.Printf("open: %v\n", err)
			continue
		}

		serial, err := device.SerialNumber(ctx)
		_ = device.Close()
		if err != nil {
			fmt.Printf("serial: %v\n", err)
			continue
		}

		fmt.Printf("serial: %s\n", serial)

		identity, ok := token2.Identify(serial)
		if ok {
			fmt.Printf(
				"model: revision=%s form-factor=%q branding=%q prefix=%s check=%c suffix=%s\n",
				identity.Model.Revision,
				identity.Model.FormFactor,
				identity.Model.Branding,
				identity.Prefix,
				identity.CheckDigit,
				identity.Suffix,
			)
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "token2usb:", err)
		os.Exit(1)
	}
}
