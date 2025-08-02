package core

import "fmt"

type InfraType string

const (
	InfraMongo    InfraType = "mongo"
	InfraRedis    InfraType = "redis"
	InfraPostgres InfraType = "postgres"
	InfraElastic  InfraType = "elasticsearch"
	InfraKafka    InfraType = "kafka"
)

type ProvisionInput struct {
	ID        string
	InfraType InfraType
	Metadata  map[string]string      // e.g. cluster, namespace, pod, labels
	Config    map[string]interface{} // e.g. version, size, etc.
}

type ProvisionResult struct {
	Endpoint  string
	SecretRef string
	Status    string
}

type ConnectionInfo struct {
	ID        string
	InfraType InfraType
	Endpoint  string
	Auth      map[string]string
	Meta      map[string]string
}

type InfraProvisioner interface {
	Create(input ProvisionInput) (*ProvisionResult, error)
	Destroy(input ProvisionInput) error
}

type ConnectionStore interface {
	Save(info ConnectionInfo) error
	Delete(id string) error
	Get(id string) (*ConnectionInfo, error)
}

type InfraManager struct {
	provisioners map[InfraType]InfraProvisioner
	store        ConnectionStore
}

func NewInfraManager(provs map[InfraType]InfraProvisioner, store ConnectionStore) *InfraManager {
	return &InfraManager{
		provisioners: provs,
		store:        store,
	}
}

func (m *InfraManager) CreateInfra(input ProvisionInput) (*ProvisionResult, error) {
	prov, ok := m.provisioners[input.InfraType]
	if !ok {
		return nil, fmt.Errorf("no provisioner found for type %s", input.InfraType)
	}
	res, err := prov.Create(input)
	if err != nil {
		return nil, err
	}
	conn := ConnectionInfo{
		ID:        input.ID,
		InfraType: input.InfraType,
		Endpoint:  res.Endpoint,
		Auth:      map[string]string{"secretRef": res.SecretRef},
		Meta:      map[string]string{"status": res.Status},
	}
	return res, m.store.Save(conn)
}

func (m *InfraManager) DestroyInfra(input ProvisionInput) error {
	prov, ok := m.provisioners[input.InfraType]
	if !ok {
		return fmt.Errorf("no provisioner found for type %s", input.InfraType)
	}
	if err := prov.Destroy(input); err != nil {
		return err
	}
	return m.store.Delete(input.ID)
}
