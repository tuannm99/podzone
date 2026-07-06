package backofficeclient

import (
	"testing"

	"github.com/stretchr/testify/require"

	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
)

func TestStoreOwnerID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		request  storeentity.StoreRequest
		expected string
	}{
		{
			name: "explicit owner",
			request: storeentity.StoreRequest{
				RequestedBy: "platform-admin",
				OwnerID:     "tenant-root",
			},
			expected: "tenant-root",
		},
		{
			name: "legacy request fallback",
			request: storeentity.StoreRequest{
				RequestedBy: "tenant-root",
			},
			expected: "tenant-root",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, test.expected, storeOwnerID(test.request))
		})
	}
}
