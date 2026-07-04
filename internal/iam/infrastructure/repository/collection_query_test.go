package repository

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/collection"
)

func TestBuildIAMCollectionQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		query       collection.Query
		columns     map[string]collectionColumn
		search      []string
		defaultSort string
		wantOrder   string
		wantWhere   int
	}{
		{
			name: "organization search and sort",
			query: collection.Query{
				Page:          2,
				PageSize:      10,
				Search:        `ops%_team`,
				SortBy:        "name",
				SortDirection: collection.SortAscending,
			},
			columns:     organizationCollectionColumns,
			search:      []string{"id", "slug", "name"},
			defaultSort: "created_at",
			wantOrder:   "organization.name ASC",
			wantWhere:   1,
		},
		{
			name: "policy equality filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "scope",
					Operator: collection.FilterEqual,
					Values:   []string{"platform"},
				}},
			},
			columns:     policyCollectionColumns,
			search:      []string{"scope", "name", "description"},
			defaultSort: "created_at",
			wantOrder:   "created_at DESC",
			wantWhere:   1,
		},
		{
			name: "group name contains filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "name",
					Operator: collection.FilterContains,
					Values:   []string{"operator"},
				}},
			},
			columns:     groupCollectionColumns,
			search:      []string{"scope", "name", "description"},
			defaultSort: "created_at",
			wantOrder:   "created_at DESC",
			wantWhere:   1,
		},
		{
			name: "policy version default filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "isDefault",
					Operator: collection.FilterEqual,
					Values:   []string{"true"},
				}},
				SortBy: "version",
			},
			columns:     policyVersionCollectionColumns,
			search:      []string{"pv.version"},
			defaultSort: "pv.created_at",
			wantOrder:   "pv.version DESC",
			wantWhere:   1,
		},
		{
			name: "group member search",
			query: collection.Query{
				Search: "42",
				SortBy: "userId",
			},
			columns:     groupMemberCollectionColumns,
			search:      []string{"CAST(gm.user_id AS TEXT)"},
			defaultSort: "gm.created_at",
			wantOrder:   "gm.user_id DESC",
			wantWhere:   1,
		},
		{
			name: "inline policy name filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "name",
					Operator: collection.FilterContains,
					Values:   []string{"orders"},
				}},
			},
			columns:     inlinePolicyCollectionColumns,
			search:      []string{"ip.name", "ip.description"},
			defaultSort: "ip.created_at",
			wantOrder:   "ip.created_at DESC",
			wantWhere:   1,
		},
		{
			name: "attachment type filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "attachmentType",
					Operator: collection.FilterIn,
					Values:   []string{"group", "role"},
				}},
			},
			columns:     policyAttachmentCollectionColumns,
			search:      []string{"attachment_type", "scope", "tenant_id"},
			defaultSort: "created_at",
			wantOrder:   "created_at DESC",
			wantWhere:   1,
		},
		{
			name: "tenant membership role filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "roleName",
					Operator: collection.FilterEqual,
					Values:   []string{"tenant_admin"},
				}},
			},
			columns:     tenantMembershipCollectionColumns,
			search:      []string{"CAST(tm.user_id AS TEXT)", "r.name", "tm.status"},
			defaultSort: "tm.created_at",
			wantOrder:   "tm.created_at DESC",
			wantWhere:   1,
		},
		{
			name: "tenant invite email search",
			query: collection.Query{
				Search: "ops@example.com",
				SortBy: "expiresAt",
			},
			columns:     tenantInviteCollectionColumns,
			search:      []string{"ti.email", "r.name", "ti.status"},
			defaultSort: "ti.created_at",
			wantOrder:   "ti.expires_at DESC",
			wantWhere:   1,
		},
		{
			name: "platform role status filter",
			query: collection.Query{
				Filters: []collection.Filter{{
					Field:    "status",
					Operator: collection.FilterEqual,
					Values:   []string{"active"},
				}},
				SortBy:        "roleName",
				SortDirection: collection.SortAscending,
			},
			columns:     platformRoleCollectionColumns,
			search:      []string{"r.name", "upr.status"},
			defaultSort: "upr.created_at",
			wantOrder:   "r.name ASC",
			wantWhere:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			normalized, where, orderBy, err := buildIAMCollectionQuery(
				tt.query,
				tt.columns,
				tt.search,
				tt.defaultSort,
			)
			require.NoError(t, err)
			assert.Equal(t, tt.wantOrder, orderBy)
			assert.Len(t, where, tt.wantWhere)
			assert.GreaterOrEqual(t, normalized.Page, collection.DefaultPage)

			builder := sq.Select("*").From("items")
			for _, predicate := range where {
				builder = builder.Where(predicate)
			}
			query, args, sqlErr := builder.PlaceholderFormat(sq.Dollar).ToSql()
			require.NoError(t, sqlErr)
			assert.NotEmpty(t, query)
			if tt.query.Search != "" {
				assert.NotContains(t, query, tt.query.Search)
				require.NotEmpty(t, args)
				assert.Equal(t, "%"+escapeIAMLike(tt.query.Search)+"%", args[0])
			}
		})
	}
}

func TestBuildIAMCollectionQueryRejectsUnsupportedField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildIAMCollectionQuery(
		collection.Query{
			Filters: []collection.Filter{{
				Field:    "raw_sql",
				Operator: collection.FilterEqual,
				Values:   []string{"1"},
			}},
		},
		policyCollectionColumns,
		[]string{"name"},
		"created_at",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, collection.ErrInvalidQuery)
}

func TestBuildIAMCollectionQueryRejectsTextOperatorForNonTextField(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildIAMCollectionQuery(
		collection.Query{
			Filters: []collection.Filter{{
				Field:    "isSystem",
				Operator: collection.FilterContains,
				Values:   []string{"true"},
			}},
		},
		policyCollectionColumns,
		[]string{"name"},
		"created_at",
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, collection.ErrInvalidQuery)
}
