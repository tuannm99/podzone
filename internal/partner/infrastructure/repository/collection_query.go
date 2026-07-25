package repository

import (
	sq "github.com/Masterminds/squirrel"
	partnerdomain "github.com/tuannm99/podzone/internal/partner/domain"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/collection/sqlspec"
)

// See docs/07-problems/repository-query-layer-2026-07-24.md RQ1 follow-up
// (2026-07-25 audit of auth/iam/onboarding/partner): unlike auth, iam, and
// onboarding's own collection-query builders, this one had no text-column
// guard, so `contains`/`starts_with` on slaDays/routingPriority (INTEGER)
// or createdAt/updatedAt (TIMESTAMPTZ) built invalid SQL
// (`LOWER(sla_days) LIKE ...`) and surfaced as a Postgres error rather than
// a validation error - the same defect class as RQ1.
var partnerSpec = sqlspec.Spec{
	Entity:      "partner",
	DefaultSort: "routingPriority",
	SearchOrder: []string{"code", "name", "contactName", "contactEmail", "notes", "partnerType", "status"},
	Fields: map[string]sqlspec.Field{
		"id":                  {Column: "id", Kind: sqlspec.KindText},
		"code":                {Column: "code", Kind: sqlspec.KindText},
		"name":                {Column: "name", Kind: sqlspec.KindText},
		"contactName":         {Column: "contact_name", Kind: sqlspec.KindText},
		"contactEmail":        {Column: "contact_email", Kind: sqlspec.KindText},
		"notes":               {Column: "notes", Kind: sqlspec.KindText},
		"partnerType":         {Column: "partner_type", Kind: sqlspec.KindText},
		"status":              {Column: "status", Kind: sqlspec.KindText},
		"slaDays":             {Column: "sla_days"},
		"routingPriority":     {Column: "routing_priority"},
		"baseFulfillmentCost": {Column: "base_fulfillment_cost", Kind: sqlspec.KindText},
		"createdAt":           {Column: "created_at"},
		"updatedAt":           {Column: "updated_at"},
	},
}

func buildPartnerCollectionQuery(
	query partnerdomain.ListPartnersQuery,
) (collection.Query, []sq.Sqlizer, string, error) {
	base := []sq.Sqlizer{sq.Eq{"tenant_id": query.TenantID}}
	if query.Status != "" {
		base = append(base, sq.Eq{"status": query.Status})
	}
	if query.PartnerType != "" {
		base = append(base, sq.Eq{"partner_type": query.PartnerType})
	}
	return partnerSpec.Compile(query.Collection, base...)
}
