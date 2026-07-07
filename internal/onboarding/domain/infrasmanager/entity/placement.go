package entity

import "time"

type PlacementRuntime string

const (
	PlacementRuntimeLocalDocker PlacementRuntime = "local_docker"
	PlacementRuntimeDocker      PlacementRuntime = "docker"
	PlacementRuntimeKubernetes  PlacementRuntime = "kubernetes"
	PlacementRuntimeK8s         PlacementRuntime = "k8s"
	PlacementRuntimeTerraform   PlacementRuntime = "terraform"
)

type StorePlacementRequest struct {
	RequestID   string
	TenantID    string
	StoreID     string
	Subdomain   string
	RequestedBy string
}

type PlacementPlan struct {
	RequestID         string
	TenantID          string
	StoreID           string
	Runtime           PlacementRuntime
	ClusterName       string
	Mode              string
	DBName            string
	SchemaName        string
	ProviderMeta      map[string]string
	InventorySnapshot ResourceInventory
	CapacitySnapshot  CapacitySnapshot
	PolicyDecision    PlacementPolicyDecision
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PlacementAllocation struct {
	ID           string
	RequestID    string
	TenantID     string
	StoreID      string
	Runtime      PlacementRuntime
	ClusterName  string
	Mode         string
	DBName       string
	SchemaName   string
	Endpoint     string
	SecretRef    string
	Status       string
	ProviderMeta map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PlacementRoute struct {
	ClusterName string
	Mode        string
	DBName      string
	SchemaName  string
}
