package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type InfraManager struct {
	provisioners map[InfraType]InfraProvisioner
	store        ConnectionStore

	// Consul publishing is async via outbox, so manager doesn't need consul directly.
}

func NewInfraManager(provs map[InfraType]InfraProvisioner, store ConnectionStore) *InfraManager {
	return &InfraManager{provisioners: provs, store: store}
}

func (m *InfraManager) CreateInfra(input ProvisionInput) (*ProvisionResult, error) {
	prov, ok := m.provisioners[input.InfraType]
	if !ok {
		return nil, fmt.Errorf("no provisioner found for type %s", input.InfraType)
	}

	if input.Name == "" {
		input.Name = "default"
	}

	evID := uuid.NewString()
	now := time.Now()

	_ = m.store.AppendEvent(ConnectionEvent{
		EventID:   evID,
		TenantID:  input.TenantID,
		Name:      input.Name,
		InfraType: input.InfraType,
		Action:    "create",
		Status:    "started",
		Request: map[string]interface{}{
			"id":        input.ID,
			"metadata":  input.Metadata,
			"config":    input.Config,
			"infraType": string(input.InfraType),
		},
		CreatedAt: now,
	})

	res, err := prov.Create(input)
	if err != nil {
		_ = m.store.AppendEvent(ConnectionEvent{
			EventID:   evID,
			TenantID:  input.TenantID,
			Name:      input.Name,
			InfraType: input.InfraType,
			Action:    "create",
			Status:    "failed",
			Error:     err.Error(),
			CreatedAt: time.Now(),
		})
		return nil, err
	}

	conn := ConnectionInfo{
		TenantID:  input.TenantID,
		Name:      input.Name,
		InfraType: input.InfraType,
		Endpoint:  res.Endpoint,
		SecretRef: res.SecretRef,
		Status:    res.Status,
		Meta:      input.Metadata,
		Config:    input.Config,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	if err := m.store.Upsert(conn); err != nil {
		_ = m.store.AppendEvent(ConnectionEvent{
			EventID:   evID,
			TenantID:  input.TenantID,
			Name:      input.Name,
			InfraType: input.InfraType,
			Action:    "create",
			Status:    "failed",
			Error:     "save connection failed: " + err.Error(),
			CreatedAt: time.Now(),
		})
		return nil, err
	}

	_ = m.store.AppendEvent(ConnectionEvent{
		EventID:   evID,
		TenantID:  input.TenantID,
		Name:      input.Name,
		InfraType: input.InfraType,
		Action:    "create",
		Status:    "succeeded",
		Result: map[string]interface{}{
			"endpoint":   res.Endpoint,
			"secretRef":  res.SecretRef,
			"status":     res.Status,
			"updated_at": time.Now().UTC().Format(time.RFC3339),
		},
		CreatedAt: time.Now(),
	})

	// Enqueue outbox for Consul publish (runtime snapshot)
	consulKey := buildConsulKey(input.TenantID, input.InfraType, input.Name)
	consulVal, _ := json.Marshal(map[string]interface{}{
		"endpoint":  res.Endpoint,
		"secretRef": res.SecretRef,
		"status":    res.Status,
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
		"tenantID":  input.TenantID,
		"infraType": string(input.InfraType),
		"name":      input.Name,
		"meta":      input.Metadata,
		"config":    input.Config,
	})

	_ = m.store.EnqueueOutbox(OutboxMessage{
		EventID:    evID,
		Topic:      "consul.publish",
		Payload:    map[string]interface{}{"key": consulKey, "value": string(consulVal)},
		Status:     "pending",
		RetryCount: 0,
		NextRetry:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})

	return res, nil
}

func (m *InfraManager) DestroyInfra(input ProvisionInput) error {
	prov, ok := m.provisioners[input.InfraType]
	if !ok {
		return fmt.Errorf("no provisioner found for type %s", input.InfraType)
	}
	if input.Name == "" {
		input.Name = "default"
	}

	evID := uuid.NewString()
	_ = m.store.AppendEvent(ConnectionEvent{
		EventID:   evID,
		TenantID:  input.TenantID,
		Name:      input.Name,
		InfraType: input.InfraType,
		Action:    "destroy",
		Status:    "started",
		CreatedAt: time.Now(),
	})

	if err := prov.Destroy(input); err != nil {
		_ = m.store.AppendEvent(ConnectionEvent{
			EventID:   evID,
			TenantID:  input.TenantID,
			Name:      input.Name,
			InfraType: input.InfraType,
			Action:    "destroy",
			Status:    "failed",
			Error:     err.Error(),
			CreatedAt: time.Now(),
		})
		return err
	}

	if err := m.store.SoftDelete(input.TenantID, input.InfraType, input.Name); err != nil {
		return err
	}

	// Enqueue consul delete
	consulKey := buildConsulKey(input.TenantID, input.InfraType, input.Name)
	_ = m.store.EnqueueOutbox(OutboxMessage{
		EventID:   evID,
		Topic:     "consul.delete",
		Payload:   map[string]interface{}{"key": consulKey},
		Status:    "pending",
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	_ = m.store.AppendEvent(ConnectionEvent{
		EventID:   evID,
		TenantID:  input.TenantID,
		Name:      input.Name,
		InfraType: input.InfraType,
		Action:    "destroy",
		Status:    "succeeded",
		CreatedAt: time.Now(),
	})
	return nil
}

func buildConsulKey(tenantID string, infraType InfraType, name string) string {
	return fmt.Sprintf("podzone/tenants/%s/connections/%s/%s", tenantID, infraType, name)
}
