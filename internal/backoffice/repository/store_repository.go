package repository

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/tuannm99/podzone/pkg/contextfx"
	"github.com/tuannm99/podzone/pkg/postgresfx"
	"github.com/tuannm99/podzone/internal/backoffice/models"
)

var (
	ErrStoreNotFound = errors.New("store not found")
	ErrUnauthorized  = errors.New("unauthorized access")
)

// StoreRepository handles store data persistence
type StoreRepository struct {
	dbManager *postgresfx.TenantDBManager
	logger    *zap.Logger
}

// NewStoreRepository creates a new store repository
func NewStoreRepository(dbManager *postgresfx.TenantDBManager, logger *zap.Logger) *StoreRepository {
	return &StoreRepository{
		dbManager: dbManager,
		logger:    logger,
	}
}

// Create creates a new store
func (r *StoreRepository) Create(ctx context.Context, store *models.Store) error {
	tenantID, err := contextfx.GetTenantID(ctx)
	if err != nil {
		return ErrUnauthorized
	}
	store.OwnerID = tenantID

	db, err := r.dbManager.GetDB(ctx)
	if err != nil {
		return err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&models.Store{}); err != nil {
		return err
	}

	return db.Create(store).Error
}

// GetByID gets a store by ID
func (r *StoreRepository) GetByID(ctx context.Context, id string) (*models.Store, error) {
	tenantID, err := contextfx.GetTenantID(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	db, err := r.dbManager.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	var store models.Store
	err = db.Where("id = ? AND owner_id = ?", id, tenantID).First(&store).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &store, nil
}

// List lists all stores for a tenant
func (r *StoreRepository) List(ctx context.Context) ([]*models.Store, error) {
	tenantID, err := contextfx.GetTenantID(ctx)
	if err != nil {
		return nil, ErrUnauthorized
	}

	db, err := r.dbManager.GetDB(ctx)
	if err != nil {
		return nil, err
	}

	var stores []*models.Store
	err = db.Where("owner_id = ?", tenantID).Find(&stores).Error
	if err != nil {
		return nil, err
	}

	return stores, nil
}

// Update updates a store
func (r *StoreRepository) Update(ctx context.Context, store *models.Store) error {
	tenantID, err := contextfx.GetTenantID(ctx)
	if err != nil {
		return ErrUnauthorized
	}

	db, err := r.dbManager.GetDB(ctx)
	if err != nil {
		return err
	}

	result := db.Model(&models.Store{}).
		Where("id = ? AND owner_id = ?", store.ID, tenantID).
		Updates(store)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrStoreNotFound
	}

	return nil
}

// Delete deletes a store
func (r *StoreRepository) Delete(ctx context.Context, id string) error {
	tenantID, err := contextfx.GetTenantID(ctx)
	if err != nil {
		return ErrUnauthorized
	}

	db, err := r.dbManager.GetDB(ctx)
	if err != nil {
		return err
	}

	result := db.Where("id = ? AND owner_id = ?", id, tenantID).Delete(&models.Store{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrStoreNotFound
	}

	return nil
}
