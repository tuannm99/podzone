package sqlrow

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"
)

type sampleRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Ignored   string `db:"-"`
	untagged  string //nolint:unused // exercises the "no db tag" skip path
	CreatedAt string `db:"created_at"`
}

func TestColumnsOf(t *testing.T) {
	got := ColumnsOf[sampleRow]()
	require.Equal(t, []string{"id", "name", "created_at"}, got)

	// Cached path returns the same content on a second call.
	require.Equal(t, []string{"id", "name", "created_at"}, ColumnsOf[sampleRow]())
}

func TestSelectFrom(t *testing.T) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, _, err := SelectFrom[sampleRow](psql, "samples").Where(sq.Eq{"id": "1"}).ToSql()
	require.NoError(t, err)
	require.Equal(t, `SELECT id, name, created_at FROM samples WHERE id = $1`, query)
}
