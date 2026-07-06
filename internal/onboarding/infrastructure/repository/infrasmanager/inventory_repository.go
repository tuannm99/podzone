package repository

import (
	"context"
	"fmt"
	"time"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func (s *MongoStore) ListDatabaseClusters(
	ctx context.Context,
	query collection.Query,
) (collection.Page[entity.DatabaseCluster], error) {
	normalized, filter, sort, err := buildDatabaseClusterCollection(query)
	if err != nil {
		return collection.Page[entity.DatabaseCluster]{}, err
	}
	total, err := s.dbCol.CountDocuments(ctx, filter)
	if err != nil {
		return collection.Page[entity.DatabaseCluster]{}, err
	}
	cursor, err := s.dbCol.Find(
		ctx,
		filter,
		options.Find().
			SetSort(sort).
			SetSkip(int64(normalized.Offset())).
			SetLimit(int64(normalized.PageSize)),
	)
	if err != nil {
		return collection.Page[entity.DatabaseCluster]{}, err
	}
	defer cursor.Close(ctx)

	var docs []databaseClusterDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return collection.Page[entity.DatabaseCluster]{}, err
	}
	items := make([]entity.DatabaseCluster, 0, len(docs))
	for _, doc := range docs {
		items = append(items, doc.toEntity())
	}
	return collection.NewPage(items, total, normalized), nil
}

func (s *MongoStore) UpsertDatabaseCluster(ctx context.Context, cluster entity.DatabaseCluster) error {
	now := time.Now().UTC()
	_, err := s.dbCol.UpdateOne(
		ctx,
		bson.M{"name": cluster.Name},
		bson.M{
			"$set": bson.M{
				"engine":              cluster.Engine,
				"region":              cluster.Region,
				"placement_db":        cluster.PlacementDB,
				"max_tenants":         cluster.MaxTenants,
				"current_tenants":     cluster.CurrentTenants,
				"max_schemas":         cluster.MaxSchemas,
				"current_schemas":     cluster.CurrentSchemas,
				"max_connections":     cluster.MaxConnections,
				"current_connections": cluster.CurrentConnections,
				"status":              cluster.Status,
				"healthy":             cluster.Healthy,
				"updated_at":          now,
			},
			"$setOnInsert": bson.M{"name": cluster.Name, "created_at": now},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *MongoStore) DeleteDatabaseCluster(ctx context.Context, name string) error {
	return archiveResource(ctx, s.dbCol, name)
}

func (s *MongoStore) ListKubernetesClusters(
	ctx context.Context,
	query collection.Query,
) (collection.Page[entity.KubernetesCluster], error) {
	normalized, filter, sort, err := buildKubernetesClusterCollection(query)
	if err != nil {
		return collection.Page[entity.KubernetesCluster]{}, err
	}
	total, err := s.k8sCol.CountDocuments(ctx, filter)
	if err != nil {
		return collection.Page[entity.KubernetesCluster]{}, err
	}
	cursor, err := s.k8sCol.Find(
		ctx,
		filter,
		options.Find().
			SetSort(sort).
			SetSkip(int64(normalized.Offset())).
			SetLimit(int64(normalized.PageSize)),
	)
	if err != nil {
		return collection.Page[entity.KubernetesCluster]{}, err
	}
	defer cursor.Close(ctx)

	var docs []kubernetesClusterDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return collection.Page[entity.KubernetesCluster]{}, err
	}
	items := make([]entity.KubernetesCluster, 0, len(docs))
	for _, doc := range docs {
		items = append(items, doc.toEntity())
	}
	return collection.NewPage(items, total, normalized), nil
}

func (s *MongoStore) UpsertKubernetesCluster(ctx context.Context, cluster entity.KubernetesCluster) error {
	now := time.Now().UTC()
	namespaces := make([]kubernetesNamespaceDoc, 0, len(cluster.Namespaces))
	for _, namespace := range cluster.Namespaces {
		namespaces = append(namespaces, kubernetesNamespaceDoc{
			Name:           namespace.Name,
			MaxTenants:     namespace.MaxTenants,
			CurrentTenants: namespace.CurrentTenants,
			CPUMilli:       namespace.CPUMilli,
			MemoryMi:       namespace.MemoryMi,
			Status:         namespace.Status,
			Healthy:        namespace.Healthy,
		})
	}
	_, err := s.k8sCol.UpdateOne(
		ctx,
		bson.M{"name": cluster.Name},
		bson.M{
			"$set": bson.M{
				"region":     cluster.Region,
				"namespaces": namespaces,
				"status":     cluster.Status,
				"healthy":    cluster.Healthy,
				"updated_at": now,
			},
			"$setOnInsert": bson.M{"name": cluster.Name, "created_at": now},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *MongoStore) DeleteKubernetesCluster(ctx context.Context, name string) error {
	return archiveResource(ctx, s.k8sCol, name)
}

func (s *MongoStore) ListRuntimePools(
	ctx context.Context,
	query collection.Query,
) (collection.Page[entity.RuntimePool], error) {
	normalized, filter, sort, err := buildRuntimePoolCollection(query)
	if err != nil {
		return collection.Page[entity.RuntimePool]{}, err
	}
	total, err := s.poolCol.CountDocuments(ctx, filter)
	if err != nil {
		return collection.Page[entity.RuntimePool]{}, err
	}
	cursor, err := s.poolCol.Find(
		ctx,
		filter,
		options.Find().
			SetSort(sort).
			SetSkip(int64(normalized.Offset())).
			SetLimit(int64(normalized.PageSize)),
	)
	if err != nil {
		return collection.Page[entity.RuntimePool]{}, err
	}
	defer cursor.Close(ctx)

	var docs []runtimePoolDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return collection.Page[entity.RuntimePool]{}, err
	}
	items := make([]entity.RuntimePool, 0, len(docs))
	for _, doc := range docs {
		items = append(items, doc.toEntity())
	}
	return collection.NewPage(items, total, normalized), nil
}

func (s *MongoStore) UpsertRuntimePool(ctx context.Context, pool entity.RuntimePool) error {
	now := time.Now().UTC()
	_, err := s.poolCol.UpdateOne(
		ctx,
		bson.M{"name": pool.Name},
		bson.M{
			"$set": bson.M{
				"kind":            pool.Kind,
				"max_tenants":     pool.MaxTenants,
				"current_tenants": pool.CurrentTenants,
				"status":          pool.Status,
				"healthy":         pool.Healthy,
				"updated_at":      now,
			},
			"$setOnInsert": bson.M{"name": pool.Name, "created_at": now},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *MongoStore) DeleteRuntimePool(ctx context.Context, name string) error {
	return archiveResource(ctx, s.poolCol, name)
}

func archiveResource(ctx context.Context, resourceCollection *mongo.Collection, name string) error {
	result, err := resourceCollection.UpdateOne(
		ctx,
		bson.M{"name": name, "status": bson.M{"$ne": "archived"}},
		bson.M{"$set": bson.M{"status": "archived", "healthy": false, "updated_at": time.Now().UTC()}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return entity.ErrResourceNotFound
	}
	return nil
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
		Status:             d.Status,
		Healthy:            d.Healthy,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
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
		Status:     d.Status,
		Healthy:    d.Healthy,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
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
		Status:         d.Status,
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
		Status:         d.Status,
		Healthy:        d.Healthy,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
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
