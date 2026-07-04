package repository

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	"github.com/tuannm99/podzone/pkg/collection"
)

type OrganizationRepositoryImpl struct {
	db *sqlx.DB
}

var _ outputport.OrganizationRepository = (*OrganizationRepositoryImpl)(nil)

func NewOrganizationRepository(p repoParams) outputport.OrganizationRepository {
	return &OrganizationRepositoryImpl{db: p.DB}
}

func (r *OrganizationRepositoryImpl) Create(
	ctx context.Context,
	org entity.Organization,
) (*entity.Organization, error) {
	var out organizationModel
	if err := r.db.GetContext(
		ctx,
		&out,
		`INSERT INTO iam_organizations (id, slug, name, root_user_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, slug, name, root_user_id, created_at, updated_at`,
		org.ID, org.Slug, org.Name, org.RootUserID, org.CreatedAt, org.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &entity.Organization{
		ID:         out.ID,
		Slug:       out.Slug,
		Name:       out.Name,
		RootUserID: out.RootUserID,
		CreatedAt:  out.CreatedAt,
		UpdatedAt:  out.UpdatedAt,
	}, nil
}

func (r *OrganizationRepositoryImpl) EnsureRoot(
	ctx context.Context,
	org entity.Organization,
) (*entity.Organization, error) {
	var out organizationModel
	err := r.db.GetContext(
		ctx,
		&out,
		`WITH root_org AS (
			INSERT INTO iam_organizations (id, slug, name, root_user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (root_user_id) WHERE root_user_id > 0
			DO UPDATE SET root_user_id = EXCLUDED.root_user_id
			RETURNING id, slug, name, root_user_id, created_at, updated_at
		), attached_tenants AS (
			UPDATE tenants t
			SET org_id = (SELECT id FROM root_org), updated_at = now()
			WHERE t.org_id = ''
			  AND EXISTS (
				SELECT 1
				FROM tenant_memberships tm
				JOIN iam_roles r ON r.id = tm.role_id
				WHERE tm.tenant_id = t.id
				  AND tm.user_id = $4
				  AND tm.status = 'active'
				  AND r.name = 'tenant_owner'
			  )
			RETURNING t.id
		), root_membership AS (
			INSERT INTO iam_organization_memberships (
				org_id, user_id, role_id, status, created_at, updated_at
			)
			SELECT root_org.id, root_org.root_user_id, role.id, 'active', now(), now()
			FROM root_org
			JOIN iam_roles role ON role.name = 'organization_root'
			ON CONFLICT (org_id, user_id) DO UPDATE
			SET role_id = EXCLUDED.role_id, status = 'active', updated_at = now()
			RETURNING org_id
		)
		SELECT id, slug, name, root_user_id, created_at, updated_at
		FROM root_org`,
		org.ID,
		org.Slug,
		org.Name,
		org.RootUserID,
		org.CreatedAt,
		org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return organizationModelToEntity(out), nil
}

func (r *OrganizationRepositoryImpl) UpsertMembership(
	ctx context.Context,
	membership entity.OrganizationMembership,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO iam_organization_memberships (
			org_id, user_id, role_id, status, created_at, updated_at
		 )
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (org_id, user_id) DO UPDATE
		 SET role_id = EXCLUDED.role_id,
		     status = EXCLUDED.status,
		     updated_at = EXCLUDED.updated_at`,
		membership.OrgID,
		membership.UserID,
		membership.RoleID,
		membership.Status,
		membership.CreatedAt,
		membership.UpdatedAt,
	)
	return err
}

func (r *OrganizationRepositoryImpl) DeleteMembership(ctx context.Context, orgID string, userID uint) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_organization_memberships WHERE org_id = $1 AND user_id = $2`,
		orgID,
		userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return entity.ErrOrganizationMembershipNotFound
	}
	return nil
}

func (r *OrganizationRepositoryImpl) GetMembership(
	ctx context.Context,
	orgID string,
	userID uint,
) (*entity.OrganizationMembership, error) {
	var row organizationMembershipModel
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT om.org_id, om.user_id, om.role_id, r.name AS role_name,
		        om.status, om.created_at, om.updated_at
		 FROM iam_organization_memberships om
		 JOIN iam_roles r ON r.id = om.role_id
		 WHERE om.org_id = $1 AND om.user_id = $2`,
		orgID,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrOrganizationMembershipNotFound
		}
		return nil, err
	}
	return organizationMembershipModelToEntity(row), nil
}

func (r *OrganizationRepositoryImpl) ListMemberships(
	ctx context.Context,
	orgID string,
	query collection.Query,
) (collection.Page[entity.OrganizationMembership], error) {
	page, err := listIAMCollectionModels[organizationMembershipModel](
		ctx,
		r.db,
		query,
		"iam_organization_memberships om JOIN iam_roles r ON r.id = om.role_id",
		[]string{
			"om.org_id",
			"om.user_id",
			"om.role_id",
			"r.name AS role_name",
			"om.status",
			"om.created_at",
			"om.updated_at",
		},
		[]sq.Sqlizer{sq.Eq{"om.org_id": orgID}},
		organizationMembershipCollectionColumns,
		[]string{"CAST(om.user_id AS TEXT)", "r.name", "om.status"},
		"om.created_at",
		"om.user_id ASC",
	)
	if err != nil {
		return collection.Page[entity.OrganizationMembership]{}, err
	}
	items := make([]entity.OrganizationMembership, 0, len(page.Items))
	for _, row := range page.Items {
		items = append(items, *organizationMembershipModelToEntity(row))
	}
	return collection.NewPage(items, page.Total, query), nil
}

func organizationMembershipModelToEntity(row organizationMembershipModel) *entity.OrganizationMembership {
	return &entity.OrganizationMembership{
		OrgID:     row.OrgID,
		UserID:    row.UserID,
		RoleID:    row.RoleID,
		RoleName:  row.RoleName,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func (r *OrganizationRepositoryImpl) List(
	ctx context.Context,
	queryArg collection.Query,
) (collection.Page[entity.Organization], error) {
	return r.list(ctx, queryArg, nil)
}

func (r *OrganizationRepositoryImpl) ListByUserID(
	ctx context.Context,
	userID uint,
	queryArg collection.Query,
) (collection.Page[entity.Organization], error) {
	return r.list(ctx, queryArg, &userID)
}

func (r *OrganizationRepositoryImpl) list(
	ctx context.Context,
	queryArg collection.Query,
	memberUserID *uint,
) (collection.Page[entity.Organization], error) {
	table := "iam_organizations organization"
	normalized, where, orderBy, err := buildIAMCollectionQuery(
		queryArg,
		organizationCollectionColumns,
		[]string{"organization.id", "organization.slug", "organization.name"},
		"organization.created_at",
	)
	if err != nil {
		return collection.Page[entity.Organization]{}, err
	}
	if memberUserID != nil {
		table += " JOIN iam_organization_memberships membership ON membership.org_id = organization.id"
		where = append(
			where,
			sq.Eq{
				"membership.user_id": *memberUserID,
				"membership.status":  entity.MembershipStatusActive,
			},
		)
	}
	countBuilder := sq.Select("COUNT(*)").From(table)
	for _, predicate := range where {
		countBuilder = countBuilder.Where(predicate)
	}
	countSQL, countArgs, err := countBuilder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return collection.Page[entity.Organization]{}, err
	}
	var total int64
	if err := r.db.GetContext(ctx, &total, countSQL, countArgs...); err != nil {
		return collection.Page[entity.Organization]{}, err
	}

	builder := sq.Select(
		"organization.id",
		"organization.slug",
		"organization.name",
		"organization.root_user_id",
		"organization.created_at",
		"organization.updated_at",
	).From(table)
	for _, predicate := range where {
		builder = builder.Where(predicate)
	}
	listSQL, listArgs, err := builder.
		OrderBy(orderBy, "organization.id ASC").
		Limit(uint64(normalized.PageSize)).
		Offset(uint64(normalized.Offset())).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return collection.Page[entity.Organization]{}, err
	}
	var rows []organizationModel
	if err := r.db.SelectContext(ctx, &rows, listSQL, listArgs...); err != nil {
		return collection.Page[entity.Organization]{}, err
	}
	out := make([]entity.Organization, 0, len(rows))
	for _, row := range rows {
		out = append(out, entity.Organization{
			ID:         row.ID,
			Slug:       row.Slug,
			Name:       row.Name,
			RootUserID: row.RootUserID,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		})
	}
	return collection.NewPage(out, total, normalized), nil
}

func (r *OrganizationRepositoryImpl) GetByID(ctx context.Context, orgID string) (*entity.Organization, error) {
	var row organizationModel
	if err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, slug, name, root_user_id, created_at, updated_at
		 FROM iam_organizations
		 WHERE id = $1`,
		orgID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrOrganizationNotFound
		}
		return nil, err
	}
	return organizationModelToEntity(row), nil
}

func (r *OrganizationRepositoryImpl) GetByRootUserID(
	ctx context.Context,
	userID uint,
) (*entity.Organization, error) {
	var row organizationModel
	if err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, slug, name, root_user_id, created_at, updated_at
		 FROM iam_organizations
		 WHERE root_user_id = $1`,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrOrganizationNotFound
		}
		return nil, err
	}
	return organizationModelToEntity(row), nil
}

func (r *OrganizationRepositoryImpl) IsRoot(ctx context.Context, orgID string, userID uint) (bool, error) {
	var exists bool
	err := r.db.GetContext(
		ctx,
		&exists,
		`SELECT EXISTS(
			SELECT 1 FROM iam_organizations WHERE id = $1 AND root_user_id = $2
		)`,
		orgID,
		userID,
	)
	return exists, err
}

func organizationModelToEntity(row organizationModel) *entity.Organization {
	return &entity.Organization{
		ID:         row.ID,
		Slug:       row.Slug,
		Name:       row.Name,
		RootUserID: row.RootUserID,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func (r *OrganizationRepositoryImpl) AttachServiceControlPolicy(
	ctx context.Context,
	orgID string,
	policyID uint64,
) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO iam_org_service_control_policies (org_id, policy_id, created_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (org_id, policy_id) DO NOTHING`,
		orgID, policyID,
	)
	return err
}

func (r *OrganizationRepositoryImpl) DetachServiceControlPolicy(
	ctx context.Context,
	orgID string,
	policyID uint64,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM iam_org_service_control_policies WHERE org_id = $1 AND policy_id = $2`,
		orgID,
		policyID,
	)
	return err
}

func (r *OrganizationRepositoryImpl) ListServiceControlPolicies(
	ctx context.Context,
	orgID string,
) ([]entity.Policy, error) {
	var rows []policyModel
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT p.id, p.scope, p.name, p.description, p.is_system, p.default_version, p.created_at, p.updated_at
		 FROM iam_org_service_control_policies osp
		 JOIN iam_policies p ON p.id = osp.policy_id
		 WHERE osp.org_id = $1
		 ORDER BY osp.created_at ASC, p.name ASC`,
		orgID,
	); err != nil {
		return nil, err
	}
	out := make([]entity.Policy, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toEntity())
	}
	return out, nil
}

func (r *OrganizationRepositoryImpl) ListServiceControlPolicyStatements(
	ctx context.Context,
	orgID string,
) ([]entity.PolicyStatement, error) {
	var rows []policyStatementModel
	if err := r.db.SelectContext(ctx, &rows,
		`SELECT ps.id, ps.policy_id, p.name AS policy_name, ps.effect, ps.action_pattern, ps.resource_pattern, ps.conditions_json, ps.created_at
		 FROM iam_org_service_control_policies osp
		 JOIN iam_policies p ON p.id = osp.policy_id
		 JOIN iam_policy_statements ps ON ps.policy_id = p.id
		 WHERE osp.org_id = $1
		 ORDER BY ps.id ASC`,
		orgID,
	); err != nil {
		return nil, err
	}
	return toPolicyStatements(rows), nil
}
