package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	catalogctx "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	"github.com/tuannm99/podzone/internal/backoffice/migrations"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Repository struct {
	mgr pdtenantdb.Manager
}

var _ catalogctx.ProductSetupRepository = (*Repository)(nil)

func New(mgr pdtenantdb.Manager) catalogctx.ProductSetupRepository {
	return &Repository{mgr: mgr}
}

type productSetupDraftRow struct {
	ID          string    `db:"id"`
	StoreID     string    `db:"store_id"`
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
	StoreID              string    `db:"store_id"`
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

func (r *Repository) ListDrafts(ctx context.Context, storeID string) ([]catalogctx.ProductSetupDraft, error) {
	query, args, err := psql.
		Select(
			"id",
			"store_id",
			"name",
			"partner",
			"base_cost",
			"retail_price",
			"status",
			"notes",
			"created_at",
			"updated_at",
		).
		From("product_setup_drafts").
		Where(sq.Eq{"store_id": storeID}).
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

	out := make([]catalogctx.ProductSetupDraft, 0, len(rows))
	for _, row := range rows {
		out = append(out, catalogctx.ProductSetupDraft(row))
	}
	return out, nil
}

func (r *Repository) GetDraftByID(ctx context.Context, storeID, id string) (*catalogctx.ProductSetupDraft, error) {
	query, args, err := psql.
		Select(
			"id",
			"store_id",
			"name",
			"partner",
			"base_cost",
			"retail_price",
			"status",
			"notes",
			"created_at",
			"updated_at",
		).
		From("product_setup_drafts").
		Where(sq.Eq{"id": id, "store_id": storeID}).
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

	out := catalogctx.ProductSetupDraft(row)
	return &out, nil
}

func (r *Repository) CreateDraft(
	ctx context.Context,
	draft catalogctx.ProductSetupDraft,
) (*catalogctx.ProductSetupDraft, error) {
	query, args, err := psql.
		Insert("product_setup_drafts").
		Columns("id", "store_id", "name", "partner", "base_cost", "retail_price", "status", "notes", "created_at", "updated_at").
		Values(draft.ID, draft.StoreID, draft.Name, draft.Partner, draft.BaseCost, draft.RetailPrice, draft.Status, draft.Notes, draft.CreatedAt, draft.UpdatedAt).
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

func (r *Repository) ListCandidates(ctx context.Context, storeID string) ([]catalogctx.ProductSetupCandidate, error) {
	query, args, err := psql.
		Select(
			"id", "store_id", "draft_id", "title", "sku", "partner", "base_cost", "retail_price",
			"estimated_margin", "status", "channel", "variants_json", "artwork_checklist_json",
			"merchandising_notes", "updated_at",
		).
		From("product_setup_candidates").
		Where(sq.Eq{"store_id": storeID}).
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

	out := make([]catalogctx.ProductSetupCandidate, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapCandidateRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, mapped)
	}
	return out, nil
}

func (r *Repository) GetCandidateByID(
	ctx context.Context,
	storeID string,
	id string,
) (*catalogctx.ProductSetupCandidate, error) {
	query, args, err := psql.
		Select(
			"id", "store_id", "draft_id", "title", "sku", "partner", "base_cost", "retail_price",
			"estimated_margin", "status", "channel", "variants_json", "artwork_checklist_json",
			"merchandising_notes", "updated_at",
		).
		From("product_setup_candidates").
		Where(sq.Eq{"id": id, "store_id": storeID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var row productSetupCandidateRow
	if err := r.withTenantTx(ctx, func(tx *sqlx.Tx) error {
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
	}); err != nil {
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

func (r *Repository) GetCandidateByDraftID(
	ctx context.Context,
	storeID string,
	draftID string,
) (*catalogctx.ProductSetupCandidate, error) {
	query, args, err := psql.
		Select(
			"id", "store_id", "draft_id", "title", "sku", "partner", "base_cost", "retail_price",
			"estimated_margin", "status", "channel", "variants_json", "artwork_checklist_json",
			"merchandising_notes", "updated_at",
		).
		From("product_setup_candidates").
		Where(sq.Eq{"draft_id": draftID, "store_id": storeID}).
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

func (r *Repository) CreateCandidate(
	ctx context.Context,
	candidate catalogctx.ProductSetupCandidate,
) (*catalogctx.ProductSetupCandidate, error) {
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
		Columns("id", "store_id", "draft_id", "title", "sku", "partner", "base_cost", "retail_price", "estimated_margin", "status", "channel", "variants_json", "artwork_checklist_json", "merchandising_notes", "updated_at").
		Values(candidate.ID, candidate.StoreID, candidate.DraftID, candidate.Title, candidate.SKU, candidate.Partner, candidate.BaseCost, candidate.RetailPrice, candidate.EstimatedMargin, candidate.Status, candidate.Channel, string(variantsJSON), string(checklistJSON), candidate.MerchandisingNotes, candidate.UpdatedAt).
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

func (r *Repository) UpdateCandidateStatus(
	ctx context.Context,
	storeID string,
	id, status string,
) (*catalogctx.ProductSetupCandidate, error) {
	query, args, err := psql.
		Update("product_setup_candidates").
		Set("status", status).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": id, "store_id": storeID}).
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

	candidates, err := r.ListCandidates(ctx, storeID)
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

func (r *Repository) withTenantTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return err
	}
	return r.mgr.WithTenantTx(ctx, tenantID, nil, fn)
}

func ensureProductSetupTables(ctx context.Context, tx *sqlx.Tx) error {
	return migrations.ApplyTx(ctx, tx)
}

func mapCandidateRow(row productSetupCandidateRow) (catalogctx.ProductSetupCandidate, error) {
	var variants []catalogctx.ProductSetupVariant
	if err := json.Unmarshal([]byte(row.VariantsJSON), &variants); err != nil {
		return catalogctx.ProductSetupCandidate{}, err
	}
	var checklist catalogctx.ProductSetupArtworkChecklist
	if err := json.Unmarshal([]byte(row.ArtworkChecklistJSON), &checklist); err != nil {
		return catalogctx.ProductSetupCandidate{}, err
	}
	return catalogctx.ProductSetupCandidate{
		ID:                 row.ID,
		StoreID:            row.StoreID,
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
