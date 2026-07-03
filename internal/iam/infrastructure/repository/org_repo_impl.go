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

func (r *OrganizationRepositoryImpl) List(
	ctx context.Context,
	queryArg collection.Query,
) (collection.Page[entity.Organization], error) {
	normalized, where, orderBy, err := buildIAMCollectionQuery(
		queryArg,
		organizationCollectionColumns,
		[]string{"id", "slug", "name"},
		"created_at",
	)
	if err != nil {
		return collection.Page[entity.Organization]{}, err
	}
	countBuilder := sq.Select("COUNT(*)").From("iam_organizations")
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

	builder := sq.Select("id", "slug", "name", "root_user_id", "created_at", "updated_at").
		From("iam_organizations")
	for _, predicate := range where {
		builder = builder.Where(predicate)
	}
	listSQL, listArgs, err := builder.
		OrderBy(orderBy, "id ASC").
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
