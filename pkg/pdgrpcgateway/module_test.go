package pdgrpcgateway

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestGatewayRegistersStandardErrorDetails(t *testing.T) {
	t.Parallel()

	messageType, err := protoregistry.GlobalTypes.FindMessageByURL(
		"type.googleapis.com/google.rpc.ErrorInfo",
	)

	require.NoError(t, err)
	require.Equal(t, "google.rpc.ErrorInfo", string(messageType.Descriptor().FullName()))
}
