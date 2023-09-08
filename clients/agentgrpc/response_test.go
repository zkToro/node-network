package agentgrpc

import (
	"testing"

	"zktoro/zktoro-core-go/protocol"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	r := require.New(t)

	errMsg1 := "operation failed"
	errMsg2 := "deadline exceeded"

	r.EqualError(Error([]*protocol.Error{
		{
			Message: errMsg1,
		},
		{
			Message: errMsg2,
		},
	}), "operation failed, deadline exceeded")

	r.EqualError(Error([]*protocol.Error{}), "<empty error list>")
}
