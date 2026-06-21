package provider

import (
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func runtimePoolName(runtime entity.PlacementRuntime, namespace string) string {
	switch runtime {
	case entity.PlacementRuntimeKubernetes, entity.PlacementRuntimeK8s:
		return "k8s/" + toolkit.FirstNonEmpty(namespace, "default")
	case entity.PlacementRuntimeTerraform:
		return "terraform/default"
	default:
		return "docker/default"
	}
}

func selectDatabaseCluster(inventory entity.ResourceInventory, name string) *entity.DatabaseCluster {
	for i := range inventory.DBClusters {
		if inventory.DBClusters[i].Name == name {
			return &inventory.DBClusters[i]
		}
	}
	for i := range inventory.DBClusters {
		if inventory.DBClusters[i].Healthy {
			return &inventory.DBClusters[i]
		}
	}
	return nil
}

func selectNamespace(inventory entity.ResourceInventory, name string) *entity.KubernetesNamespace {
	for clusterIdx := range inventory.K8sClusters {
		for namespaceIdx := range inventory.K8sClusters[clusterIdx].Namespaces {
			namespace := &inventory.K8sClusters[clusterIdx].Namespaces[namespaceIdx]
			if namespace.Name == name {
				return namespace
			}
		}
	}
	for clusterIdx := range inventory.K8sClusters {
		if !inventory.K8sClusters[clusterIdx].Healthy {
			continue
		}
		for namespaceIdx := range inventory.K8sClusters[clusterIdx].Namespaces {
			namespace := &inventory.K8sClusters[clusterIdx].Namespaces[namespaceIdx]
			if namespace.Healthy {
				return namespace
			}
		}
	}
	return nil
}

func selectRuntimePool(
	inventory entity.ResourceInventory,
	name string,
	runtime entity.PlacementRuntime,
) *entity.RuntimePool {
	for i := range inventory.RuntimePools {
		if inventory.RuntimePools[i].Name == name {
			return &inventory.RuntimePools[i]
		}
	}
	for i := range inventory.RuntimePools {
		if inventory.RuntimePools[i].Kind == string(runtime) && inventory.RuntimePools[i].Healthy {
			return &inventory.RuntimePools[i]
		}
	}
	for i := range inventory.RuntimePools {
		if inventory.RuntimePools[i].Healthy {
			return &inventory.RuntimePools[i]
		}
	}
	return nil
}

func appendDBCapacityReasons(snapshot *entity.CapacitySnapshot, cluster entity.DatabaseCluster) {
	if !cluster.Healthy {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "database cluster is unhealthy")
	}
	if cluster.MaxTenants <= 0 || cluster.MaxSchemas <= 0 || cluster.MaxConnections <= 0 {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "database cluster capacity is unknown")
	}
	if cluster.MaxTenants > 0 && cluster.CurrentTenants >= cluster.MaxTenants {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "database cluster tenant capacity exceeded")
	}
	if cluster.MaxSchemas > 0 && cluster.CurrentSchemas >= cluster.MaxSchemas {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "database schema capacity exceeded")
	}
	if cluster.MaxConnections > 0 && cluster.CurrentConnections >= cluster.MaxConnections {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "database connection capacity exceeded")
	}
}

func appendNamespaceCapacityReasons(snapshot *entity.CapacitySnapshot, namespace entity.KubernetesNamespace) {
	if !namespace.Healthy {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "kubernetes namespace is unhealthy")
	}
	if namespace.MaxTenants <= 0 {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "kubernetes namespace capacity is unknown")
	}
	if namespace.MaxTenants > 0 && namespace.CurrentTenants >= namespace.MaxTenants {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "kubernetes namespace tenant capacity exceeded")
	}
}

func appendRuntimePoolCapacityReasons(snapshot *entity.CapacitySnapshot, pool entity.RuntimePool) {
	if !pool.Healthy {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "runtime pool is unhealthy")
	}
	if pool.MaxTenants <= 0 {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "runtime pool capacity is unknown")
	}
	if pool.MaxTenants > 0 && pool.CurrentTenants >= pool.MaxTenants {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "runtime pool tenant capacity exceeded")
	}
}
