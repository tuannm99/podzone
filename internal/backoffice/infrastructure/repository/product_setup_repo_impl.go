package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type ProductSetupRepositoryImpl struct {
	mgr pdtenantdb.Manager
}

func NewProductSetupRepository(mgr pdtenantdb.Manager) outputport.ProductSetupRepository {
	return &ProductSetupRepositoryImpl{mgr: mgr}
}

const ensureProductSetupDraftsSQL = `
CREATE TABLE IF NOT EXISTS product_setup_drafts (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	partner TEXT NOT NULL,
	base_cost TEXT NOT NULL,
	retail_price TEXT NOT NULL,
	status TEXT NOT NULL,
	notes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
)`

const ensureProductSetupCandidatesSQL = `
CREATE TABLE IF NOT EXISTS product_setup_candidates (
	id TEXT PRIMARY KEY,
	draft_id TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	sku TEXT NOT NULL,
	partner TEXT NOT NULL,
	base_cost TEXT NOT NULL,
	retail_price TEXT NOT NULL,
	estimated_margin TEXT NOT NULL,
	status TEXT NOT NULL,
	channel TEXT NOT NULL,
	variants_json TEXT NOT NULL,
	artwork_checklist_json TEXT NOT NULL,
	merchandising_notes TEXT NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ NOT NULL
)`

type productSetupDraftRow struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Partner     string    `db:"partner"`
	BaseCost    string    `db:"base_cost"`
	RetailPrice string    `db:"retail_price"`
	Status      string    `db:"status"`
	Notes       string    `db:"notes"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type productSetupCandidateRow struct {
	ID                   string    `db:"id"`
	DraftID              string    `db:"draft_id"`
	Title                string    `db:"title"`
	SKU                  string    `db:"sku"`
	Partner              string    `db:"partner"`
	BaseCost             string    `db:"base_cost"`
	RetailPrice          string    `db:"retail_price"`
	EstimatedMargin      string    `db:"estimated_margin"`
	Status               string    `db:"status"`
	Channel              string    `db:"channel"`
	VariantsJSON         string    `db:"variants_json"`
	ArtworkChecklistJSON string    `db:"artwork_checklist_json"`
	MerchandisingNotes   string    `db:"merchandising_notes"`
	UpdatedAt            time.Time `db:"updated_at"`
}

func (r *ProductSetupRepositoryImpl) ListDrafts(ctx context.Context) ([]entity.ProductSetupDraft, error) {
	query, args, err := psql.
		Select("id", "name", "partner", "base_cost", "retail_price", "status", "notes", "created_at", "updated_at").
		From("product_setup_drafts").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []productSetupDraftRow
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureProductSetupTables(ctx, tx); err != nil {
			return err
		}
		return tx.SelectContext(ctx, &rows, query, args...)
	}); err != nil {
		return nil, err
	}

	out := make([]entity.ProductSetupDraft, 0, len(rows))
	for _, row := range rows {
		out = append(out, entity.ProductSetupDraft(row))
	}
	return out, nil
}

func (r *ProductSetupRepositoryImpl) GetDraftByID(ctx context.Context, id string) (*entity.ProductSetupDraft, error) {
	query, args, err := psql.
		Select("id", "name", "partner", "base_cost", "retail_price", "status", "notes", "created_at", "updated_at").
		From("product_setup_drafts").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var row productSetupDraftRow
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureProductSetupTables(ctx, tx); err != nil {
			return err
		}
		if err := tx.GetContext(ctx, &row, query, args...); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("product setup draft not found")
			}
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	out := entity.ProductSetupDraft(row)
	return &out, nil
}

func (r *ProductSetupRepositoryImpl) CreateDraft(ctx context.Context, draft entity.ProductSetupDraft) (*entity.ProductSetupDraft, error) {
	query, args, err := psql.
		Insert("product_setup_drafts").
		Columns("id", "name", "partner", "base_cost", "retail_price", "status", "notes", "created_at", "updated_at").
		Values(draft.ID, draft.Name, draft.Partner, draft.BaseCost, draft.RetailPrice, draft.Status, draft.Notes, draft.CreatedAt, draft.UpdatedAt).
		ToSql()
	if err != nil {
		return nil, err
	}

	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureProductSetupTables(ctx, tx); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, query, args...)
		return err
	}); err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *ProductSetupRepositoryImpl) ListCandidates(ctx context.Context) ([]entity.ProductSetupCandidate, error) {
	query, args, err := psql.
		Select("id", "draft_id", "title", "sku", "partner", "base_cost", "retail_price", "estimated_margin", "status", "channel", "variants_json", "artwork_checklist_json", "merchandising_notes", "updated_at").
		From("product_setup_candidates").
		OrderBy("updated_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []productSetupCandidateRow
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureProductSetupTables(ctx, tx); err != nil {
			return err
		}
		return tx.SelectContext(ctx, &rows, query, args...)
	}); err != nil {
		return nil, err
	}

	out := make([]entity.ProductSetupCandidate, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapCandidateRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

func (r *ProductSetupRepositoryImpl) GetCandidateByDraftID(ctx context.Context, draftID string) (*entity.ProductSetupCandidate, error) {
	query, args, err := psql.
		Select("id", "draft_id", "title", "sku", "partner", "base_cost", "retail_price", "estimated_margin", "status", "channel", "variants_json", "artwork_checklist_json", "merchandising_notes", "updated_at").
		From("product_setup_candidates").
		Where(sq.Eq{"draft_id": draftID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var row productSetupCandidateRow
	err = r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureProductSetupTables(ctx, tx); err != nil {
			return err
		}
		if err := tx.GetContext(ctx, &row, query, args...); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, nil
	}
	mapped, err := mapCandidateRow(row)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

func (r *ProductSetupRepositoryImpl) CreateCandidate(ctx context.Context, candidate entity.ProductSetupCandidate) (*entity.ProductSetupCandidate, error) {
	variantsJSON, err := json.Marshal(candidate.Variants)
	if err != nil {
		return nil, err
	}
	checklistJSON, err := json.Marshal(candidate.ArtworkChecklist)
	if err != nil {
		return nil, err
	}

	query, args, err := psql.
		Insert("product_setup_candidates").
		Columns("id", "draft_id", "title", "sku", "partner", "base_cost", "retail_price", "estimated_margin", "status", "channel", "variants_json", "artwork_checklist_json", "merchandising_notes", "updated_at").
		Values(candidate.ID, candidate.DraftID, candidate.Title, candidate.SKU, candidate.Partner, candidate.BaseCost, candidate.RetailPrice, candidate.EstimatedMargin, candidate.Status, candidate.Channel, string(variantsJSON), string(checklistJSON), candidate.MerchandisingNotes, candidate.UpdatedAt).
		ToSql()
	if err != nil {
		return nil, err
	}

	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureProductSetupTables(ctx, tx); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, query, args...)
		return err
	}); err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (r *ProductSetupRepositoryImpl) UpdateCandidateStatus(ctx context.Context, id, status string) (*entity.ProductSetupCandidate, error) {
	query, args, err := psql.
		Update("product_setup_candidates").
		Set("status", status).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
		if err := ensureProductSetupTables(ctx, tx); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("product setup candidate not found")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	candidates, err := r.ListCandidates(ctx)
	if err != nil {
		return nil, err
	}
	for i := range candidates {
		if candidates[i].ID == id {
			return &candidates[i], nil
		}
	}
	return nil, fmt.Errorf("product setup candidate not found")
}

func (r *ProductSetupRepositoryImpl) withTenantTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return err
	}
	return r.mgr.WithTenantTx(ctx, tenantID, nil, fn)
}

func ensureProductSetupTables(ctx context.Context, tx *sqlx.Tx) error {
	if _, err := tx.ExecContext(ctx, ensureProductSetupDraftsSQL); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, ensureProductSetupCandidatesSQL)
	return err
}

func mapCandidateRow(row productSetupCandidateRow) (entity.ProductSetupCandidate, error) {
	var variants []entity.ProductSetupVariant
	if err := json.Unmarshal([]byte(row.VariantsJSON), &variants); err != nil {
		return entity.ProductSetupCandidate{}, err
	}
	var checklist entity.ProductSetupArtworkChecklist
	if err := json.Unmarshal([]byte(row.ArtworkChecklistJSON), &checklist); err != nil {
		return entity.ProductSetupCandidate{}, err
	}
	return entity.ProductSetupCandidate{
		ID:                 row.ID,
		DraftID:            row.DraftID,
		Title:              row.Title,
		SKU:                row.SKU,
		Partner:            row.Partner,
		BaseCost:           row.BaseCost,
		RetailPrice:        row.RetailPrice,
		EstimatedMargin:    row.EstimatedMargin,
		Status:             row.Status,
		Channel:            row.Channel,
		UpdatedAt:          row.UpdatedAt,
		Variants:           variants,
		ArtworkChecklist:   checklist,
		MerchandisingNotes: row.MerchandisingNotes,
	}, nil
}
