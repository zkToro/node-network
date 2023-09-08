package agentgrpc

import (
	"errors"

	"zktoro/zktoro-core-go/protocol"
)

// Error makes single error from our common errors defined in protobuf.
func Error(respErrs []*protocol.Error) error {
	var errMsg string
	for i, respErr := range respErrs {
		if i > 0 {
			errMsg += ", "
		}
		errMsg += respErr.Message
	}
	if len(errMsg) == 0 {
		return errors.New("<empty error list>")
	}
	return errors.New(errMsg)
}
