package infrasmanager

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
)

type Service struct {
	st core.ConnectionStore
}

func NewService(st core.ConnectionStore) *Service {
	return &Service{st: st}
}

// ManualUpsertConnection stores connection and enqueues publish snapshot to Consul.
func (s *Service) ManualUpsertConnection(
	tenantID string,
	req UpsertConnectionRequest,
	actor map[string]string,
) (*UpsertConnectionResponse, error) {
	name := req.Name
	if name == "" {
		name = "default"
	}

	corrID := uuid.NewString()

	_ = s.st.AppendEvent(core.ConnectionEvent{
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

	conn := core.ConnectionInfo{
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

	if err := s.st.Upsert(conn); err != nil {
		_ = s.st.AppendEvent(core.ConnectionEvent{
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

	consulKey := core.BuildConsulKey(tenantID, req.InfraType, name)

	snap := map[string]interface{}{
		"tenantID":  tenantID,
		"infraType": string(req.InfraType),
		"name":      name,
		"endpoint":  req.Endpoint,
		"secretRef": req.SecretRef,
		"status":    firstNonEmpty(req.Status, "active"),
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
		"meta":      req.Meta,
		"config":    req.Config,
	}
	valBytes, err := json.Marshal(snap)
	if err != nil {
		_ = s.st.AppendEvent(core.ConnectionEvent{
			ID:            uuid.NewString(),
			CorrelationID: corrID,
			TenantID:      tenantID,
			InfraType:     req.InfraType,
			Name:          name,
			Action:        "manual_upsert",
			Status:        "failed",
			Error:         "marshal consul snapshot failed: " + err.Error(),
			Actor:         actor,
			CreatedAt:     time.Now(),
		})
		return nil, err
	}

	_ = s.st.EnqueueOutbox(core.OutboxMessage{
		EventID:       uuid.NewString(),
		CorrelationID: corrID,
		Topic:         "consul.publish",
		Payload:       map[string]interface{}{"key": consulKey, "value": string(valBytes)},
		TenantID:      tenantID,
		InfraType:     req.InfraType,
		Name:          name,
		Status:        "pending",
		RetryCount:    0,
		NextRetry:     time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	_ = s.st.AppendEvent(core.ConnectionEvent{
		ID:            uuid.NewString(),
		CorrelationID: corrID,
		TenantID:      tenantID,
		InfraType:     req.InfraType,
		Name:          name,
		Action:        "manual_upsert",
		Status:        "succeeded",
		Result: map[string]interface{}{
			"consul_key": consulKey,
			"queued":     true,
		},
		Actor:     actor,
		CreatedAt: time.Now(),
	})

	// Fetch latest to return (optional)
	latest, _ := s.st.Get(tenantID, req.InfraType, name)
	if latest == nil {
		latest = &conn
	}

	return &UpsertConnectionResponse{
		CorrelationID: corrID,
		Connection:    toDTO(*latest),
		Queued:        true,
		ConsulKey:     consulKey,
	}, nil
}

func (s *Service) DeleteConnection(
	tenantID string,
	infraType core.InfraType,
	name string,
	actor map[string]string,
) (string, error) {
	if name == "" {
		name = "default"
	}

	corrID := uuid.NewString()

	_ = s.st.AppendEvent(core.ConnectionEvent{
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

	if err := s.st.SoftDelete(tenantID, infraType, name); err != nil {
		_ = s.st.AppendEvent(core.ConnectionEvent{
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

	consulKey := core.BuildConsulKey(tenantID, infraType, name)
	_ = s.st.EnqueueOutbox(core.OutboxMessage{
		EventID:       uuid.NewString(),
		CorrelationID: corrID,
		Topic:         "consul.delete",
		Payload:       map[string]interface{}{"key": consulKey},
		TenantID:      tenantID,
		InfraType:     infraType,
		Name:          name,
		Status:        "pending",
		NextRetry:     time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	_ = s.st.AppendEvent(core.ConnectionEvent{
		ID:            uuid.NewString(),
		CorrelationID: corrID,
		TenantID:      tenantID,
		InfraType:     infraType,
		Name:          name,
		Action:        "manual_delete",
		Status:        "succeeded",
		Result:        map[string]interface{}{"consul_key": consulKey, "queued": true},
		Actor:         actor,
		CreatedAt:     time.Now(),
	})

	return corrID, nil
}

func (s *Service) GetConnection(tenantID string, infraType core.InfraType, name string) (*ConnectionDTO, error) {
	c, err := s.st.Get(tenantID, infraType, name)
	if err != nil {
		return nil, err
	}
	dto := toDTO(*c)
	return &dto, nil
}

func (s *Service) ListConnections(
	tenantID string,
	infraType core.InfraType,
	includeDeleted bool,
	limit, offset int,
) ([]ConnectionDTO, error) {
	items, err := s.st.ListConnections(tenantID, infraType, includeDeleted, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]ConnectionDTO, 0, len(items))
	for _, it := range items {
		out = append(out, toDTO(it))
	}
	return out, nil
}

func (s *Service) ListEvents(
	tenantID string,
	infraType core.InfraType,
	name string,
	correlationID string,
	limit, offset int,
) ([]ConnectionEventDTO, error) {
	items, err := s.st.ListEvents(tenantID, infraType, name, correlationID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]ConnectionEventDTO, 0, len(items))
	for _, it := range items {
		out = append(out, toEventDTO(it))
	}
	return out, nil
}

func firstNonEmpty(v string, d string) string {
	if v == "" {
		return d
	}
	return v
}

