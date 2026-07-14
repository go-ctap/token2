// Package apdu implements the ISO/IEC 7816-4 command and response APDU subset
// used by Token2 devices.
package apdu

import (
	"context"
	"fmt"
)

const (
	// StatusSuccess is the ISO/IEC 7816 success status word.
	StatusSuccess uint16 = 0x9000

	statusMoreData    = 0x61
	statusMoreDataAlt = 0x9f

	instructionGetResponse = 0xc0
)

// Command describes a command APDU. The encoder supports the cases currently
// used by Token2: a four-byte header with optional short or extended data.
type Command struct {
	CLA byte
	INS byte
	P1  byte
	P2  byte

	Data     []byte
	Extended bool
}

// Bytes encodes the command APDU.
func (c Command) Bytes() []byte {
	command := []byte{c.CLA, c.INS, c.P1, c.P2}

	if len(c.Data) == 0 {
		return command
	}

	if c.Extended {
		command = append(command, 0, byte(len(c.Data)>>8), byte(len(c.Data)))
	} else {
		command = append(command, byte(len(c.Data)))
	}

	return append(command, c.Data...)
}

// Response is a parsed response APDU.
type Response struct {
	Data []byte
	SW   uint16
}

// OK reports whether the response status is 9000.
func (r Response) OK() bool {
	return r.SW == StatusSuccess
}

// StatusError reports a non-success APDU status for a logical operation.
type StatusError struct {
	Operation string
	SW        uint16
}

func (e *StatusError) Error() string {
	if e.Operation == "" {
		return fmt.Sprintf("APDU status %04x", e.SW)
	}
	return fmt.Sprintf("%s: APDU status %04x", e.Operation, e.SW)
}

// Err returns nil for a successful response and a StatusError otherwise.
func (r Response) Err(operation string) error {
	if r.OK() {
		return nil
	}
	return &StatusError{Operation: operation, SW: r.SW}
}

// ParseResponse separates response data from the trailing status word.
func ParseResponse(response []byte) (Response, error) {
	if len(response) < 2 {
		return Response{}, fmt.Errorf("APDU response is %d bytes", len(response))
	}

	dataLength := len(response) - 2

	return Response{
		Data: response[:dataLength],
		SW:   uint16(response[dataLength])<<8 | uint16(response[dataLength+1]),
	}, nil
}

// Transceiver sends one encoded command APDU and returns one encoded response
// APDU.
type Transceiver interface {
	Transmit(context.Context, []byte) ([]byte, error)
}

// Exchange sends command and follows ISO GET RESPONSE status words until the
// device returns a final response.
func Exchange(ctx context.Context, card Transceiver, command Command) (Response, error) {
	responseBytes, err := card.Transmit(ctx, command.Bytes())
	if err != nil {
		return Response{}, err
	}

	response, err := ParseResponse(responseBytes)
	if err != nil {
		return Response{}, err
	}

	data := response.Data

	for response.SW>>8 == statusMoreData || response.SW>>8 == statusMoreDataAlt {
		le := byte(response.SW)

		responseBytes, err = card.Transmit(ctx, []byte{command.CLA, instructionGetResponse, 0, 0, le})
		if err != nil {
			return Response{}, err
		}

		response, err = ParseResponse(responseBytes)
		if err != nil {
			return Response{}, err
		}

		data = append(data, response.Data...)
	}

	response.Data = data

	return response, nil
}
