package apdu

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseErr(t *testing.T) {
	assert.NoError(t, (Response{SW: StatusSuccess}).Err("read configuration"))

	err := (Response{SW: 0x6a82}).Err("select application")
	var statusErr *StatusError
	require.True(t, errors.As(err, &statusErr))
	assert.Equal(t, "select application", statusErr.Operation)
	assert.Equal(t, uint16(0x6a82), statusErr.SW)
}

type scriptedCard struct {
	requests  [][]byte
	responses [][]byte
}

func (c *scriptedCard) Transmit(_ context.Context, command []byte) ([]byte, error) {
	c.requests = append(c.requests, append([]byte(nil), command...))
	response := c.responses[0]
	c.responses = c.responses[1:]
	return response, nil
}

func TestExchangeGetResponsePreservesCLA(t *testing.T) {
	for _, cla := range []byte{0x00, 0x80} {
		t.Run(fmt.Sprintf("CLA_%02x", cla), func(t *testing.T) {
			card := &scriptedCard{responses: [][]byte{{0xaa, 0x61, 0x02}, {0xbb, 0xcc, 0x90, 0x00}}}
			response, err := Exchange(t.Context(), card, Command{CLA: cla, INS: 0x33, Data: []byte{1, 2}})
			require.NoError(t, err)
			assert.Equal(t, []byte{0xaa, 0xbb, 0xcc}, response.Data)
			assert.Equal(t, uint16(0x9000), response.SW)
			assert.Equal(t, []byte{cla, 0xc0, 0x00, 0x00, 0x02}, card.requests[1])
		})
	}
}
