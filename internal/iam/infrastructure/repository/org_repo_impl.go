package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	entity "github.com/tuannm99/podzone/internal/iam/entity"
	"github.com/tuannm99/podzone/internal/iam/outputport"
)

type OrganizationRepositoryImpl struct {
	db *sqlx.DB
}

var _ outputport.OrganizationRepository = (*OrganizationRepositoryImpl)(nil)

func NewOrganizationRepository(p repoParams) outputport.OrganizationRepository {
	return &OrganizationRepositoryImpl{db: p.DB}
}

func (r *OrganizationRepositoryImpl) Create(ctx context.Context, org entity.Organization) (*entity.Organization, error) {
	var out organizationModel
	if err := r.db.GetContext(
		ctx,
		&out,
		`INSERT INTO iam_organizations (id, slug, name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, slug, name, created_at, updated_at`,
		org.ID, org.Slug, org.Name, org.CreatedAt, org.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &entity.Organization{
		ID:        out.ID,
		Slug:      out.Slug,
		Name:      out.Name,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}, nil
}

func (r *OrganizationRepositoryImpl) List(ctx context.Context) ([]entity.Organization, error) {
	var rows []organizationModel
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, slug, name, created_at, updated_at FROM iam_organizations ORDER BY created_at ASC, id ASC`); err != nil {
		return nil, err
	}
	out := make([]entity.Organization, 0, len(rows))
	for _, row := range rows {
		out = append(out, entity.Organization{
			ID:        row.ID,
			Slug:      row.Slug,
			Name:      row.Name,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *OrganizationRepositoryImpl) GetByID(ctx context.Context, orgID string) (*entity.Organization, error) {
	var row organizationModel
	if err := r.db.GetContext(ctx, &row, `SELECT id, slug, name, created_at, updated_at FROM iam_organizations WHERE id = $1`, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrOrganizationNotFound
		}
		return nil, err
	}
	return &entity.Organization{
		ID:        row.ID,
		Slug:      row.Slug,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *OrganizationRepositoryImpl) AttachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO iam_org_service_control_policies (org_id, policy_id, created_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (org_id, policy_id) DO NOTHING`,
		orgID, policyID,
	)
	return err
}

func (r *OrganizationRepositoryImpl) DetachServiceControlPolicy(ctx context.Context, orgID string, policyID uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM iam_org_service_control_policies WHERE org_id = $1 AND policy_id = $2`, orgID, policyID)
	return err
}

func (r *OrganizationRepositoryImpl) ListServiceControlPolicies(ctx context.Context, orgID string) ([]entity.Policy, error) {
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

func (r *OrganizationRepositoryImpl) ListServiceControlPolicyStatements(ctx context.Context, orgID string) ([]entity.PolicyStatement, error) {
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
