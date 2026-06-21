package repository

import (
	"context"
	"fmt"
	"time"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoStore) LoadResourceInventory(
	ctx context.Context,
	_ entity.StorePlacementRequest,
) (entity.ResourceInventory, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dbClusters, err := s.listDatabaseClusters(ctx)
	if err != nil {
		return entity.ResourceInventory{}, err
	}
	k8sClusters, err := s.listKubernetesClusters(ctx)
	if err != nil {
		return entity.ResourceInventory{}, err
	}
	runtimePools, err := s.listRuntimePools(ctx)
	if err != nil {
		return entity.ResourceInventory{}, err
	}
	if len(dbClusters) == 0 || len(k8sClusters) == 0 || len(runtimePools) == 0 {
		return entity.ResourceInventory{}, fmt.Errorf("resource inventory is not configured")
	}

	return entity.ResourceInventory{
		Environment:  "mongo",
		DBClusters:   dbClusters,
		K8sClusters:  k8sClusters,
		RuntimePools: runtimePools,
	}, nil
}

func (s *MongoStore) EnsureConfiguredResourceInventory(
	ctx context.Context,
	cfg onboardingconfig.StoreProvisioningConfig,
) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exists, err := s.hasResourceInventory(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	namespace := toolkit.FirstNonEmpty(cfg.KubernetesNamespace, "default")
	runtime := normalizeConfigRuntime(cfg.Runtime)
	now := time.Now().UTC()

	if _, err := s.dbCol.UpdateOne(
		ctx,
		bson.M{"name": toolkit.FirstNonEmpty(cfg.ClusterName, "pg-default")},
		bson.M{"$setOnInsert": databaseClusterDoc{
			Name:           toolkit.FirstNonEmpty(cfg.ClusterName, "pg-default"),
			Engine:         "postgres",
			PlacementDB:    toolkit.FirstNonEmpty(cfg.DBName, "podzone_tenants"),
			MaxTenants:     cfg.MaxTenantsPerDBCluster,
			MaxSchemas:     cfg.MaxSchemasPerDatabase,
			MaxConnections: cfg.MaxConnectionsPerDBCluster,
			Status:         "active",
			Healthy:        true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}},
		options.Update().SetUpsert(true),
	); err != nil {
		return err
	}

	if _, err := s.k8sCol.UpdateOne(
		ctx,
		bson.M{"name": toolkit.FirstNonEmpty(cfg.ClusterName, "pg-default")},
		bson.M{"$setOnInsert": kubernetesClusterDoc{
			Name:    toolkit.FirstNonEmpty(cfg.ClusterName, "pg-default"),
			Status:  "active",
			Healthy: true,
			Namespaces: []kubernetesNamespaceDoc{
				{
					Name:       namespace,
					MaxTenants: cfg.MaxTenantsPerNamespace,
					CPUMilli:   cfg.NamespaceCPUMilli,
					MemoryMi:   cfg.NamespaceMemoryMi,
					Status:     "active",
					Healthy:    true,
				},
			},
			CreatedAt: now,
			UpdatedAt: now,
		}},
		options.Update().SetUpsert(true),
	); err != nil {
		return err
	}

	_, err = s.poolCol.UpdateOne(
		ctx,
		bson.M{"name": runtimePoolName(runtime, namespace)},
		bson.M{"$setOnInsert": runtimePoolDoc{
			Name:       runtimePoolName(runtime, namespace),
			Kind:       string(runtime),
			MaxTenants: cfg.RuntimePoolCapacity,
			Status:     "active",
			Healthy:    true,
			CreatedAt:  now,
			UpdatedAt:  now,
		}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *MongoStore) hasResourceInventory(ctx context.Context) (bool, error) {
	dbCount, err := s.dbCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		return false, err
	}
	k8sCount, err := s.k8sCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		return false, err
	}
	poolCount, err := s.poolCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		return false, err
	}
	return dbCount > 0 || k8sCount > 0 || poolCount > 0, nil
}

func (s *MongoStore) listDatabaseClusters(ctx context.Context) ([]entity.DatabaseCluster, error) {
	cur, err := s.dbCol.Find(ctx, bson.M{"status": "active"})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := []entity.DatabaseCluster{}
	for cur.Next(ctx) {
		var doc databaseClusterDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toEntity())
	}
	return out, cur.Err()
}

func (s *MongoStore) listKubernetesClusters(ctx context.Context) ([]entity.KubernetesCluster, error) {
	cur, err := s.k8sCol.Find(ctx, bson.M{"status": "active"})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := []entity.KubernetesCluster{}
	for cur.Next(ctx) {
		var doc kubernetesClusterDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toEntity())
	}
	return out, cur.Err()
}

func (s *MongoStore) listRuntimePools(ctx context.Context) ([]entity.RuntimePool, error) {
	cur, err := s.poolCol.Find(ctx, bson.M{"status": "active"})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	out := []entity.RuntimePool{}
	for cur.Next(ctx) {
		var doc runtimePoolDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		out = append(out, doc.toEntity())
	}
	return out, cur.Err()
}

type databaseClusterDoc struct {
	Name               string    `bson:"name"`
	Engine             string    `bson:"engine"`
	Region             string    `bson:"region"`
	PlacementDB        string    `bson:"placement_db"`
	MaxTenants         int       `bson:"max_tenants"`
	CurrentTenants     int       `bson:"current_tenants"`
	MaxSchemas         int       `bson:"max_schemas"`
	CurrentSchemas     int       `bson:"current_schemas"`
	MaxConnections     int       `bson:"max_connections"`
	CurrentConnections int       `bson:"current_connections"`
	Status             string    `bson:"status"`
	Healthy            bool      `bson:"healthy"`
	CreatedAt          time.Time `bson:"created_at"`
	UpdatedAt          time.Time `bson:"updated_at"`
}

func (d databaseClusterDoc) toEntity() entity.DatabaseCluster {
	return entity.DatabaseCluster{
		Name:               d.Name,
		Engine:             d.Engine,
		Region:             d.Region,
		PlacementDB:        d.PlacementDB,
		MaxTenants:         d.MaxTenants,
		CurrentTenants:     d.CurrentTenants,
		MaxSchemas:         d.MaxSchemas,
		CurrentSchemas:     d.CurrentSchemas,
		MaxConnections:     d.MaxConnections,
		CurrentConnections: d.CurrentConnections,
		Healthy:            d.Healthy,
	}
}

type kubernetesClusterDoc struct {
	Name       string                   `bson:"name"`
	Region     string                   `bson:"region"`
	Namespaces []kubernetesNamespaceDoc `bson:"namespaces"`
	Status     string                   `bson:"status"`
	Healthy    bool                     `bson:"healthy"`
	CreatedAt  time.Time                `bson:"created_at"`
	UpdatedAt  time.Time                `bson:"updated_at"`
}

func (d kubernetesClusterDoc) toEntity() entity.KubernetesCluster {
	namespaces := make([]entity.KubernetesNamespace, 0, len(d.Namespaces))
	for _, namespace := range d.Namespaces {
		if namespace.Status != "active" {
			continue
		}
		namespaces = append(namespaces, namespace.toEntity())
	}
	return entity.KubernetesCluster{
		Name:       d.Name,
		Region:     d.Region,
		Namespaces: namespaces,
		Healthy:    d.Healthy,
	}
}

type kubernetesNamespaceDoc struct {
	Name           string `bson:"name"`
	MaxTenants     int    `bson:"max_tenants"`
	CurrentTenants int    `bson:"current_tenants"`
	CPUMilli       int    `bson:"cpu_milli"`
	MemoryMi       int    `bson:"memory_mi"`
	Status         string `bson:"status"`
	Healthy        bool   `bson:"healthy"`
}

func (d kubernetesNamespaceDoc) toEntity() entity.KubernetesNamespace {
	return entity.KubernetesNamespace{
		Name:           d.Name,
		MaxTenants:     d.MaxTenants,
		CurrentTenants: d.CurrentTenants,
		CPUMilli:       d.CPUMilli,
		MemoryMi:       d.MemoryMi,
		Healthy:        d.Healthy,
	}
}

type runtimePoolDoc struct {
	Name           string    `bson:"name"`
	Kind           string    `bson:"kind"`
	MaxTenants     int       `bson:"max_tenants"`
	CurrentTenants int       `bson:"current_tenants"`
	Status         string    `bson:"status"`
	Healthy        bool      `bson:"healthy"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

func (d runtimePoolDoc) toEntity() entity.RuntimePool {
	return entity.RuntimePool{
		Name:           d.Name,
		Kind:           d.Kind,
		MaxTenants:     d.MaxTenants,
		CurrentTenants: d.CurrentTenants,
		Healthy:        d.Healthy,
	}
}

func normalizeConfigRuntime(value string) entity.PlacementRuntime {
	switch value {
	case "k8s", "kubernetes":
		return entity.PlacementRuntimeKubernetes
	case "terraform":
		return entity.PlacementRuntimeTerraform
	default:
		return entity.PlacementRuntimeLocalDocker
	}
}

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
