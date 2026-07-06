package infrasmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

var _ inputport.Usecase = (*Interactor)(nil)

type Interactor struct {
	st          storeoutputport.ConnectionStore
	placements  storeoutputport.PlacementRepository
	plans       storeoutputport.PlacementPlanRepository
	planner     storeoutputport.PlacementPlanner
	provisioner storeoutputport.StorageProvisioner
	routeReader storeoutputport.PlacementRouteReader
	routeWriter storeoutputport.PlacementRouteWriter
	inventory   storeoutputport.ResourceInventoryRepository
}

func NewInteractor(st storeoutputport.ConnectionStore) *Interactor {
	return &Interactor{st: st}
}

type InteractorParams struct {
	fx.In

	Store       storeoutputport.ConnectionStore
	Placements  storeoutputport.PlacementRepository
	Plans       storeoutputport.PlacementPlanRepository
	Planner     storeoutputport.PlacementPlanner
	Provisioner storeoutputport.StorageProvisioner
	RouteReader storeoutputport.PlacementRouteReader
	RouteWriter storeoutputport.PlacementRouteWriter
	Inventory   storeoutputport.ResourceInventoryRepository
}

func NewInteractorWithParams(p InteractorParams) *Interactor {
	return &Interactor{
		st:          p.Store,
		placements:  p.Placements,
		plans:       p.Plans,
		planner:     p.Planner,
		provisioner: p.Provisioner,
		routeReader: p.RouteReader,
		routeWriter: p.RouteWriter,
		inventory:   p.Inventory,
	}
}

func (s *Interactor) ProvisionStorePlacement(
	ctx context.Context,
	req inputport.ProvisionStorePlacementRequest,
	actor map[string]string,
) (*inputport.ProvisionStorePlacementResponse, error) {
	if s.placements == nil || s.planner == nil || s.provisioner == nil {
		return nil, fmt.Errorf("placement provisioning runtime is not configured")
	}
	if s.plans == nil {
		return nil, fmt.Errorf("placement plan repository is not configured")
	}

	placementReq := entity.StorePlacementRequest{
		RequestID:   req.RequestID,
		TenantID:    req.TenantID,
		StoreID:     req.StoreID,
		Subdomain:   req.Subdomain,
		RequestedBy: req.RequestedBy,
	}
	existing, err := s.placements.GetTenantPlacementAllocation(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return allocationResponse(existing, "", false), nil
	}

	plan, err := s.getOrCreatePlacementPlan(ctx, placementReq)
	if err != nil {
		return nil, err
	}
	allocation, err := s.provisioner.ProvisionStorePlacement(ctx, placementReq, plan)
	if err != nil {
		return nil, err
	}
	if allocation.ID == "" {
		allocation.ID = uuid.NewString()
	}
	if allocation.Status == "" {
		allocation.Status = "ready"
	}
	if allocation.CreatedAt.IsZero() {
		allocation.CreatedAt = time.Now().UTC()
	}
	allocation.UpdatedAt = time.Now().UTC()
	if err := s.placements.SavePlacementAllocation(ctx, allocation); err != nil {
		return nil, err
	}

	upsertResp, err := s.ManualUpsertConnection(ctx, req.TenantID, inputport.UpsertConnectionRequest{
		InfraType:   entity.InfraPostgres,
		Name:        "default",
		Endpoint:    allocation.Endpoint,
		SecretRef:   allocation.SecretRef,
		Status:      "active",
		ClusterName: allocation.ClusterName,
		Mode:        allocation.Mode,
		DBName:      allocation.DBName,
		SchemaName:  allocation.SchemaName,
		Meta: map[string]string{
			"placement_allocation_id": allocation.ID,
			"store_request_id":        req.RequestID,
			"store_id":                req.StoreID,
			"store_subdomain":         req.Subdomain,
			"runtime":                 string(allocation.Runtime),
		},
		Config: map[string]interface{}{
			"driver": "postgres",
			"mode":   allocation.Mode,
		},
	}, actor)
	if err != nil {
		return nil, err
	}
	return allocationResponse(&allocation, upsertResp.CorrelationID, upsertResp.Queued), nil
}

func (s *Interactor) getOrCreatePlacementPlan(
	ctx context.Context,
	req entity.StorePlacementRequest,
) (entity.PlacementPlan, error) {
	if req.RequestID != "" {
		existing, err := s.plans.GetPlacementPlanByRequestID(ctx, req.RequestID)
		if err != nil {
			return entity.PlacementPlan{}, err
		}
		if existing != nil {
			return *existing, nil
		}
	}

	plan, err := s.planner.PlanStorePlacement(ctx, req)
	if err != nil {
		return entity.PlacementPlan{}, err
	}
	if plan.RequestID == "" {
		plan.RequestID = req.RequestID
	}
	if plan.TenantID == "" {
		plan.TenantID = req.TenantID
	}
	if plan.StoreID == "" {
		plan.StoreID = req.StoreID
	}
	if err := s.plans.SavePlacementPlan(ctx, plan); err != nil {
		return entity.PlacementPlan{}, err
	}
	return plan, nil
}

func (s *Interactor) IsPlacementRouteReady(ctx context.Context, tenantID string) (bool, error) {
	if s.routeReader == nil {
		return false, nil
	}
	return s.routeReader.IsPlacementRouteReady(ctx, tenantID)
}

func (s *Interactor) EnsurePlacementRoute(ctx context.Context, tenantID string) (bool, error) {
	if s.placements == nil || s.routeWriter == nil {
		return s.IsPlacementRouteReady(ctx, tenantID)
	}

	allocation, err := s.placements.GetTenantPlacementAllocation(ctx, tenantID)
	if err != nil || allocation == nil {
		return false, err
	}
	if err := s.routeWriter.PublishPlacementRoute(ctx, tenantID, *allocation); err != nil {
		return false, err
	}
	return true, nil
}

// ManualUpsertConnection stores a connection and enqueues its runtime KV projection.
func (s *Interactor) ManualUpsertConnection(
	ctx context.Context,
	tenantID string,
	req inputport.UpsertConnectionRequest,
	actor map[string]string,
) (*inputport.UpsertConnectionResponse, error) {
	if err := validatePlacementRequest(req); err != nil {
		return nil, err
	}

	name := req.Name
	if name == "" {
		name = "default"
	}

	corrID := uuid.NewString()

	_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
		ID:            uuid.NewString(),
		CorrelationID: corrID,
		TenantID:      tenantID,
		InfraType:     req.InfraType,
		Name:          name,
		Action:        "manual_upsert",
		Status:        "started",
		Request: map[string]interface{}{
			"endpoint":   req.Endpoint,
			"secret_ref": req.SecretRef,
			"status":     req.Status,
			"meta":       req.Meta,
			"config":     req.Config,
		},
		Actor:     actor,
		CreatedAt: time.Now(),
	})

	conn := entity.ConnectionInfo{
		TenantID:  tenantID,
		InfraType: req.InfraType,
		Name:      name,
		Endpoint:  req.Endpoint,
		SecretRef: req.SecretRef,
		Status:    req.Status,
		Meta:      req.Meta,
		Config:    req.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.st.Upsert(ctx, conn); err != nil {
		_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			TenantID:      tenantID,
			InfraType:     req.InfraType,
			Name:          name,
			Action:        "manual_upsert",
			Status:        "failed",
			Error:         err.Error(),
			Actor:         actor,
			CreatedAt:     time.Now(),
		})
		return nil, err
	}

	kvStoreKey := entity.BuildKVStoreKey(tenantID, req.InfraType, name)

	snap := map[string]interface{}{
		"tenantID":  tenantID,
		"infraType": string(req.InfraType),
		"name":      name,
		"endpoint":  req.Endpoint,
		"secretRef": req.SecretRef,
		"status":    toolkit.FirstNonEmpty(req.Status, "active"),
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
		"meta":      req.Meta,
		"config":    req.Config,
	}
	valBytes, err := json.Marshal(snap)
	if err != nil {
		_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			TenantID:      tenantID,
			InfraType:     req.InfraType,
			Name:          name,
			Action:        "manual_upsert",
			Status:        "failed",
			Error:         "marshal kv store snapshot failed: " + err.Error(),
			Actor:         actor,
			CreatedAt:     time.Now(),
		})
		return nil, err
	}

	if err := s.st.EnqueueOutbox(ctx, entity.OutboxMessage{
		EventID:       uuid.NewString(),
		CorrelationID: corrID,
		Topic:         "kv_store.publish",
		Payload:       map[string]interface{}{"key": kvStoreKey, "value": string(valBytes)},
		TenantID:      tenantID,
		InfraType:     req.InfraType,
		Name:          name,
		Status:        "pending",
		RetryCount:    0,
		NextRetry:     time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}); err != nil {
		_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			TenantID:      tenantID,
			InfraType:     req.InfraType,
			Name:          name,
			Action:        "manual_upsert",
			Status:        "failed",
			Error:         "enqueue outbox failed: " + err.Error(),
			Actor:         actor,
			CreatedAt:     time.Now(),
		})
		return nil, err
	}

	// For postgres connections, also write placement routing data so that
	// pdtenantdb's KVPlacementResolver can route tenant queries.
	if req.InfraType == entity.InfraPostgres && req.ClusterName != "" {
		placementKey := "podzone/tenants/" + tenantID + "/placement"
		placementSnap := map[string]interface{}{
			"cluster_name": req.ClusterName,
			"mode":         toolkit.FirstNonEmpty(req.Mode, "schema"),
			"db_name":      req.DBName,
			"schema_name":  req.SchemaName,
		}
		if placementBytes, err := json.Marshal(placementSnap); err == nil {
			_ = s.st.EnqueueOutbox(ctx, entity.OutboxMessage{
				EventID:       uuid.NewString(),
				CorrelationID: corrID,
				Topic:         "kv_store.publish",
				Payload:       map[string]interface{}{"key": placementKey, "value": string(placementBytes)},
				TenantID:      tenantID,
				InfraType:     req.InfraType,
				Name:          name,
				Status:        "pending",
				RetryCount:    0,
				NextRetry:     time.Now(),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			})
		}
	}

	_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
		ID:            uuid.NewString(),
		CorrelationID: corrID,
		TenantID:      tenantID,
		InfraType:     req.InfraType,
		Name:          name,
		Action:        "manual_upsert",
		Status:        "succeeded",
		Result: map[string]interface{}{
			"kv_store_key": kvStoreKey,
			"queued":       true,
		},
		Actor:     actor,
		CreatedAt: time.Now(),
	})

	// Fetch latest to return (optional)
	latest, _ := s.st.Get(ctx, tenantID, req.InfraType, name)
	if latest == nil {
		latest = &conn
	}

	return &inputport.UpsertConnectionResponse{
		CorrelationID: corrID,
		Connection:    toDTO(*latest),
		Queued:        true,
		KVStoreKey:    kvStoreKey,
	}, nil
}

func (s *Interactor) DeleteConnection(
	ctx context.Context,
	tenantID string,
	infraType entity.InfraType,
	name string,
	actor map[string]string,
) (string, error) {
	if name == "" {
		name = "default"
	}

	corrID := uuid.NewString()

	_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
		ID:            uuid.NewString(),
		CorrelationID: corrID,
		TenantID:      tenantID,
		InfraType:     infraType,
		Name:          name,
		Action:        "manual_delete",
		Status:        "started",
		Actor:         actor,
		CreatedAt:     time.Now(),
	})

	if err := s.st.SoftDelete(ctx, tenantID, infraType, name); err != nil {
		_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			TenantID:      tenantID,
			InfraType:     infraType,
			Name:          name,
			Action:        "manual_delete",
			Status:        "failed",
			Error:         err.Error(),
			Actor:         actor,
			CreatedAt:     time.Now(),
		})
		return corrID, err
	}

	if infraType == entity.InfraPostgres {
		placementKey := "podzone/tenants/" + tenantID + "/placement"
		if err := s.st.EnqueueOutbox(ctx, entity.OutboxMessage{
			EventID:       uuid.NewString(),
			CorrelationID: corrID,
			Topic:         "kv_store.delete",
			Payload:       map[string]interface{}{"key": placementKey},
			TenantID:      tenantID,
			InfraType:     infraType,
			Name:          name,
			Status:        "pending",
			NextRetry:     time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}); err != nil {
			_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
				ID:            uuid.NewString(),
				CorrelationID: corrID,
				TenantID:      tenantID,
				InfraType:     infraType,
				Name:          name,
				Action:        "manual_delete",
				Status:        "failed",
				Error:         "enqueue placement delete failed: " + err.Error(),
				Actor:         actor,
				CreatedAt:     time.Now(),
			})
			return corrID, err
		}
	}

	kvStoreKey := entity.BuildKVStoreKey(tenantID, infraType, name)
	if err := s.st.EnqueueOutbox(ctx, entity.OutboxMessage{
		EventID:       uuid.NewString(),
		CorrelationID: corrID,
		Topic:         "kv_store.delete",
		Payload:       map[string]interface{}{"key": kvStoreKey},
		TenantID:      tenantID,
		InfraType:     infraType,
		Name:          name,
		Status:        "pending",
		NextRetry:     time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}); err != nil {
		_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			TenantID:      tenantID,
			InfraType:     infraType,
			Name:          name,
			Action:        "manual_delete",
			Status:        "failed",
			Error:         "enqueue outbox failed: " + err.Error(),
			Actor:         actor,
			CreatedAt:     time.Now(),
		})
		return corrID, err
	}

	_ = s.st.AppendEvent(ctx, entity.ConnectionEvent{
		ID:            uuid.NewString(),
		CorrelationID: corrID,
		TenantID:      tenantID,
		InfraType:     infraType,
		Name:          name,
		Action:        "manual_delete",
		Status:        "succeeded",
		Result:        map[string]interface{}{"kv_store_key": kvStoreKey, "queued": true},
		Actor:         actor,
		CreatedAt:     time.Now(),
	})

	return corrID, nil
}

func validatePlacementRequest(req inputport.UpsertConnectionRequest) error {
	if req.InfraType != entity.InfraPostgres || req.ClusterName == "" {
		return nil
	}

	mode := toolkit.FirstNonEmpty(req.Mode, "schema")
	switch mode {
	case "schema":
		if req.DBName == "" {
			return fmt.Errorf("db_name is required for postgres schema mode")
		}
		if req.SchemaName == "" {
			return fmt.Errorf("schema_name is required for postgres schema mode")
		}
	case "database":
		if req.DBName == "" {
			return fmt.Errorf("db_name is required for postgres database mode")
		}
	default:
		return fmt.Errorf("invalid postgres placement mode: %s", req.Mode)
	}

	return nil
}

func allocationResponse(
	allocation *entity.PlacementAllocation,
	correlationID string,
	queued bool,
) *inputport.ProvisionStorePlacementResponse {
	return &inputport.ProvisionStorePlacementResponse{
		CorrelationID: correlationID,
		AllocationID:  allocation.ID,
		Runtime:       string(allocation.Runtime),
		ClusterName:   allocation.ClusterName,
		Mode:          allocation.Mode,
		DBName:        allocation.DBName,
		SchemaName:    allocation.SchemaName,
		Endpoint:      allocation.Endpoint,
		SecretRef:     allocation.SecretRef,
		Status:        allocation.Status,
		ProviderMeta:  allocation.ProviderMeta,
		Queued:        queued,
	}
}

func (s *Interactor) GetConnection(
	ctx context.Context,
	tenantID string,
	infraType entity.InfraType,
	name string,
) (*inputport.Connection, error) {
	c, err := s.st.Get(ctx, tenantID, infraType, name)
	if err != nil {
		return nil, err
	}
	dto := toDTO(*c)
	return &dto, nil
}

func (s *Interactor) ListConnections(
	ctx context.Context,
	tenantID string,
	includeDeleted bool,
	query collection.Query,
) (collection.Page[inputport.Connection], error) {
	page, err := s.st.ListConnections(ctx, tenantID, includeDeleted, query)
	if err != nil {
		return collection.Page[inputport.Connection]{}, err
	}
	out := make([]inputport.Connection, 0, len(page.Items))
	for _, it := range page.Items {
		out = append(out, toDTO(it))
	}
	return collection.Page[inputport.Connection]{
		Items:       out,
		Total:       page.Total,
		Page:        page.Page,
		PageSize:    page.PageSize,
		TotalPages:  page.TotalPages,
		HasNext:     page.HasNext,
		HasPrevious: page.HasPrevious,
	}, nil
}

func (s *Interactor) ListEvents(
	ctx context.Context,
	tenantID string,
	query collection.Query,
) (collection.Page[inputport.ConnectionEvent], error) {
	page, err := s.st.ListEvents(ctx, tenantID, query)
	if err != nil {
		return collection.Page[inputport.ConnectionEvent]{}, err
	}
	out := make([]inputport.ConnectionEvent, 0, len(page.Items))
	for _, it := range page.Items {
		out = append(out, toEventDTO(it))
	}
	return collection.Page[inputport.ConnectionEvent]{
		Items:       out,
		Total:       page.Total,
		Page:        page.Page,
		PageSize:    page.PageSize,
		TotalPages:  page.TotalPages,
		HasNext:     page.HasNext,
		HasPrevious: page.HasPrevious,
	}, nil
}
