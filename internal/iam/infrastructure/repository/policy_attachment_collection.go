package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

const policyAttachmentCTE = `WITH attachments AS (
	SELECT 'role' AS attachment_type, '' AS scope, '' AS tenant_id,
		r.id AS role_id, r.name AS role_name, 0::bigint AS user_id,
		0::bigint AS group_id, '' AS group_name, arpa.created_at
	FROM iam_role_policy_attachments arpa
	JOIN iam_roles r ON r.id = arpa.role_id
	WHERE arpa.policy_id = ?
	UNION ALL
	SELECT 'platform_user', upa.scope, '', 0::bigint, '', upa.user_id,
		0::bigint, '', upa.created_at
	FROM iam_user_policy_attachments upa
	WHERE upa.policy_id = ?
	UNION ALL
	SELECT 'tenant_user', 'tenant', tupa.tenant_id, 0::bigint, '',
		tupa.user_id, 0::bigint, '', tupa.created_at
	FROM iam_tenant_user_policy_attachments tupa
	WHERE tupa.policy_id = ?
	UNION ALL
	SELECT 'group', g.scope, COALESCE(g.tenant_id, ''), 0::bigint, '',
		0::bigint, g.id, g.name, gpa.created_at
	FROM iam_group_policy_attachments gpa
	JOIN iam_groups g ON g.id = gpa.group_id
	WHERE gpa.policy_id = ?
	UNION ALL
	SELECT 'role_boundary', r.scope, '', r.id, r.name, 0::bigint,
		0::bigint, '', rpb.created_at
	FROM iam_role_permission_boundaries rpb
	JOIN iam_roles r ON r.id = rpb.role_id
	WHERE rpb.policy_id = ?
	UNION ALL
	SELECT 'platform_user_boundary', 'platform', '', 0::bigint, '',
		pub.user_id, 0::bigint, '', pub.created_at
	FROM iam_platform_user_permission_boundaries pub
	WHERE pub.policy_id = ?
	UNION ALL
	SELECT 'tenant_user_boundary', 'tenant', tub.tenant_id, 0::bigint, '',
		tub.user_id, 0::bigint, '', tub.created_at
	FROM iam_tenant_user_permission_boundaries tub
	WHERE tub.policy_id = ?
	UNION ALL
	SELECT 'service_control_policy', 'organization', osp.org_id, 0::bigint, '',
		0::bigint, 0::bigint, '', osp.created_at
	FROM iam_org_service_control_policies osp
	WHERE osp.policy_id = ?
)`

func (r *PolicyRepositoryImpl) listPolicyAttachmentPage(
	ctx context.Context,
	policyID uint64,
	query collection.Query,
) (collection.Page[entity.PolicyAttachment], error) {
	normalized, where, orderBy, err := buildIAMCollectionQuery(
		query,
		policyAttachmentCollectionColumns,
		[]string{
			"attachment_type",
			"scope",
			"tenant_id",
			"role_name",
			"group_name",
			"CAST(role_id AS TEXT)",
			"CAST(user_id AS TEXT)",
			"CAST(group_id AS TEXT)",
		},
		"created_at",
	)
	if err != nil {
		return collection.Page[entity.PolicyAttachment]{}, err
	}
	prefixArgs := []any{
		policyID,
		policyID,
		policyID,
		policyID,
		policyID,
		policyID,
		policyID,
		policyID,
	}
	countBuilder := sq.Select("COUNT(*)").
		Prefix(policyAttachmentCTE, prefixArgs...).
		From("attachments")
	for _, predicate := range where {
		countBuilder = countBuilder.Where(predicate)
	}
	countSQL, countArgs, err := countBuilder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return collection.Page[entity.PolicyAttachment]{}, err
	}
	var total int64
	if err := r.db.GetContext(ctx, &total, countSQL, countArgs...); err != nil {
		return collection.Page[entity.PolicyAttachment]{}, err
	}

	listBuilder := sq.Select(
		"attachment_type",
		"scope",
		"tenant_id",
		"role_id",
		"role_name",
		"user_id",
		"group_id",
		"group_name",
		"created_at",
	).Prefix(policyAttachmentCTE, prefixArgs...).From("attachments")
	for _, predicate := range where {
		listBuilder = listBuilder.Where(predicate)
	}
	listSQL, listArgs, err := listBuilder.
		OrderBy(orderBy, "attachment_type ASC").
		Limit(uint64(normalized.PageSize)).
		Offset(uint64(normalized.Offset())).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return collection.Page[entity.PolicyAttachment]{}, err
	}
	var rows []policyAttachmentModel
	if err := r.db.SelectContext(ctx, &rows, listSQL, listArgs...); err != nil {
		return collection.Page[entity.PolicyAttachment]{}, err
	}
	items := make([]entity.PolicyAttachment, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toEntity())
	}
	return collection.NewPage(items, total, normalized), nil
}
