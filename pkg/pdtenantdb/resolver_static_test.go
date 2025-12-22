package pdtenantdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	pdtenantdb "github.com/tuannm99/podzone/pkg/pdtenantdb"
)

func TestStaticPlacementResolver_SanitizesNames(t *testing.T) {
	cfg := &pdtenantdb.Config{SharedDB: "backoffice"}
	r := pdtenantdb.NewStaticPlacementResolver(cfg)

	pl, err := r.Resolve(context.Background(), "Abc-123")
	require.NoError(t, err)

	require.Equal(t, "t_abc_123", pl.SchemaName)
	require.Equal(t, "backoffice", pl.DBName)
	require.Equal(t, pdtenantdb.ModeSchema, pl.Mode)
}
