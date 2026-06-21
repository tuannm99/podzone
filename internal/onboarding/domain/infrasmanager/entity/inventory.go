package entity

type ResourceInventory struct {
	Environment  string
	DBClusters   []DatabaseCluster
	K8sClusters  []KubernetesCluster
	RuntimePools []RuntimePool
}

type DatabaseCluster struct {
	Name               string
	Engine             string
	Region             string
	PlacementDB        string
	MaxTenants         int
	CurrentTenants     int
	MaxSchemas         int
	CurrentSchemas     int
	MaxConnections     int
	CurrentConnections int
	Healthy            bool
}

type KubernetesCluster struct {
	Name       string
	Region     string
	Namespaces []KubernetesNamespace
	Healthy    bool
}

type KubernetesNamespace struct {
	Name           string
	MaxTenants     int
	CurrentTenants int
	CPUMilli       int
	MemoryMi       int
	Healthy        bool
}

type RuntimePool struct {
	Name           string
	Kind           string
	MaxTenants     int
	CurrentTenants int
	Healthy        bool
}

type CapacitySnapshot struct {
	DBClusterName string
	DatabaseName  string
	NamespaceName string
	RuntimePool   string
	CanPlace      bool
	Reasons       []string
}

type PlacementPolicyDecision struct {
	AutoApproved     bool
	ApprovalRequired bool
	Reasons          []string
}
