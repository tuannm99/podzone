package toolkit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentifier(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "lowercases", in: "Tenant-A", want: "tenant_a"},
		{name: "trims separators", in: "___Tenant A___", want: "tenant_a"},
		{name: "numeric prefix", in: "2e0df8f6", want: "tenant_2e0df8f6"},
		{name: "empty", in: "   ", want: "tenant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, Identifier(tt.in))
		})
	}
}

func TestSchemaName(t *testing.T) {
	require.Equal(t, "t_tenant_123", SchemaName("t_", "123"))
	require.Equal(t, "tenant_tenant_123", SchemaName("tenant_", "123"))
	require.Equal(t, "t_tenant", SchemaName("", ""))
}
