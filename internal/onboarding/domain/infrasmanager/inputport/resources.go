package inputport

import "time"

type DatabaseClusterResource struct {
	Name               string    `json:"name"`
	Engine             string    `json:"engine"`
	Region             string    `json:"region"`
	PlacementDB        string    `json:"placement_db"`
	MaxTenants         int       `json:"max_tenants"`
	CurrentTenants     int       `json:"current_tenants"`
	MaxSchemas         int       `json:"max_schemas"`
	CurrentSchemas     int       `json:"current_schemas"`
	MaxConnections     int       `json:"max_connections"`
	CurrentConnections int       `json:"current_connections"`
	Status             string    `json:"status"`
	Healthy            bool      `json:"healthy"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type KubernetesNamespaceResource struct {
	Name           string `json:"name"`
	MaxTenants     int    `json:"max_tenants"`
	CurrentTenants int    `json:"current_tenants"`
	CPUMilli       int    `json:"cpu_milli"`
	MemoryMi       int    `json:"memory_mi"`
	Status         string `json:"status"`
	Healthy        bool   `json:"healthy"`
}

type KubernetesClusterResource struct {
	Name       string                        `json:"name"`
	Region     string                        `json:"region"`
	Namespaces []KubernetesNamespaceResource `json:"namespaces"`
	Status     string                        `json:"status"`
	Healthy    bool                          `json:"healthy"`
	CreatedAt  time.Time                     `json:"created_at"`
	UpdatedAt  time.Time                     `json:"updated_at"`
}

type RuntimePoolResource struct {
	Name           string    `json:"name"`
	Kind           string    `json:"kind"`
	MaxTenants     int       `json:"max_tenants"`
	CurrentTenants int       `json:"current_tenants"`
	Status         string    `json:"status"`
	Healthy        bool      `json:"healthy"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DatabaseClusterHealthCheckResponse struct {
	Name               string    `json:"name"`
	Healthy            bool      `json:"healthy"`
	CurrentTenants     int       `json:"current_tenants"`
	CurrentSchemas     int       `json:"current_schemas"`
	CurrentConnections int       `json:"current_connections"`
	Message            string    `json:"message,omitempty"`
	CheckedAt          time.Time `json:"checked_at"`
}
