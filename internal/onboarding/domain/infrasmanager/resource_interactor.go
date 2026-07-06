package infrasmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	"github.com/tuannm99/podzone/pkg/collection"
)

func (s *Interactor) ListDatabaseClusters(
	ctx context.Context,
	query collection.Query,
) (collection.Page[inputport.DatabaseClusterResource], error) {
	page, err := s.inventory.ListDatabaseClusters(ctx, query)
	if err != nil {
		return collection.Page[inputport.DatabaseClusterResource]{}, err
	}
	return mapResourcePage(page, toDatabaseClusterResource), nil
}

func (s *Interactor) UpsertDatabaseCluster(
	ctx context.Context,
	resource inputport.DatabaseClusterResource,
) error {
	cluster, err := databaseClusterFromResource(resource)
	if err != nil {
		return err
	}
	return s.inventory.UpsertDatabaseCluster(ctx, cluster)
}

func (s *Interactor) DeleteDatabaseCluster(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: database cluster name is required", entity.ErrInvalidInput)
	}
	return s.inventory.DeleteDatabaseCluster(ctx, strings.TrimSpace(name))
}

func (s *Interactor) ListKubernetesClusters(
	ctx context.Context,
	query collection.Query,
) (collection.Page[inputport.KubernetesClusterResource], error) {
	page, err := s.inventory.ListKubernetesClusters(ctx, query)
	if err != nil {
		return collection.Page[inputport.KubernetesClusterResource]{}, err
	}
	return mapResourcePage(page, toKubernetesClusterResource), nil
}

func (s *Interactor) UpsertKubernetesCluster(
	ctx context.Context,
	resource inputport.KubernetesClusterResource,
) error {
	cluster, err := kubernetesClusterFromResource(resource)
	if err != nil {
		return err
	}
	return s.inventory.UpsertKubernetesCluster(ctx, cluster)
}

func (s *Interactor) DeleteKubernetesCluster(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: Kubernetes cluster name is required", entity.ErrInvalidInput)
	}
	return s.inventory.DeleteKubernetesCluster(ctx, strings.TrimSpace(name))
}

func (s *Interactor) ListRuntimePools(
	ctx context.Context,
	query collection.Query,
) (collection.Page[inputport.RuntimePoolResource], error) {
	page, err := s.inventory.ListRuntimePools(ctx, query)
	if err != nil {
		return collection.Page[inputport.RuntimePoolResource]{}, err
	}
	return mapResourcePage(page, toRuntimePoolResource), nil
}

func (s *Interactor) UpsertRuntimePool(ctx context.Context, resource inputport.RuntimePoolResource) error {
	pool, err := runtimePoolFromResource(resource)
	if err != nil {
		return err
	}
	return s.inventory.UpsertRuntimePool(ctx, pool)
}

func (s *Interactor) DeleteRuntimePool(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: runtime pool name is required", entity.ErrInvalidInput)
	}
	return s.inventory.DeleteRuntimePool(ctx, strings.TrimSpace(name))
}

func databaseClusterFromResource(
	resource inputport.DatabaseClusterResource,
) (entity.DatabaseCluster, error) {
	resource.Name = strings.TrimSpace(resource.Name)
	resource.Engine = strings.TrimSpace(resource.Engine)
	resource.PlacementDB = strings.TrimSpace(resource.PlacementDB)
	if resource.Name == "" || resource.Engine == "" || resource.PlacementDB == "" {
		return entity.DatabaseCluster{}, fmt.Errorf(
			"%w: database cluster name, engine, and placement_db are required",
			entity.ErrInvalidInput,
		)
	}
	if hasNegative(
		resource.MaxTenants,
		resource.CurrentTenants,
		resource.MaxSchemas,
		resource.CurrentSchemas,
		resource.MaxConnections,
		resource.CurrentConnections,
	) {
		return entity.DatabaseCluster{}, fmt.Errorf("%w: capacity values cannot be negative", entity.ErrInvalidInput)
	}
	return entity.DatabaseCluster{
		Name:               resource.Name,
		Engine:             resource.Engine,
		Region:             strings.TrimSpace(resource.Region),
		PlacementDB:        resource.PlacementDB,
		MaxTenants:         resource.MaxTenants,
		CurrentTenants:     resource.CurrentTenants,
		MaxSchemas:         resource.MaxSchemas,
		CurrentSchemas:     resource.CurrentSchemas,
		MaxConnections:     resource.MaxConnections,
		CurrentConnections: resource.CurrentConnections,
		Status:             activeStatus(resource.Status),
		Healthy:            resource.Healthy,
	}, nil
}

func kubernetesClusterFromResource(
	resource inputport.KubernetesClusterResource,
) (entity.KubernetesCluster, error) {
	resource.Name = strings.TrimSpace(resource.Name)
	if resource.Name == "" {
		return entity.KubernetesCluster{}, fmt.Errorf(
			"%w: Kubernetes cluster name is required",
			entity.ErrInvalidInput,
		)
	}
	namespaces := make([]entity.KubernetesNamespace, 0, len(resource.Namespaces))
	for _, namespace := range resource.Namespaces {
		namespace.Name = strings.TrimSpace(namespace.Name)
		if namespace.Name == "" {
			return entity.KubernetesCluster{}, fmt.Errorf(
				"%w: Kubernetes namespace name is required",
				entity.ErrInvalidInput,
			)
		}
		if hasNegative(
			namespace.MaxTenants,
			namespace.CurrentTenants,
			namespace.CPUMilli,
			namespace.MemoryMi,
		) {
			return entity.KubernetesCluster{}, fmt.Errorf(
				"%w: namespace capacity values cannot be negative",
				entity.ErrInvalidInput,
			)
		}
		namespaces = append(namespaces, entity.KubernetesNamespace{
			Name:           namespace.Name,
			MaxTenants:     namespace.MaxTenants,
			CurrentTenants: namespace.CurrentTenants,
			CPUMilli:       namespace.CPUMilli,
			MemoryMi:       namespace.MemoryMi,
			Status:         activeStatus(namespace.Status),
			Healthy:        namespace.Healthy,
		})
	}
	return entity.KubernetesCluster{
		Name:       resource.Name,
		Region:     strings.TrimSpace(resource.Region),
		Namespaces: namespaces,
		Status:     activeStatus(resource.Status),
		Healthy:    resource.Healthy,
	}, nil
}

func runtimePoolFromResource(resource inputport.RuntimePoolResource) (entity.RuntimePool, error) {
	resource.Name = strings.TrimSpace(resource.Name)
	resource.Kind = strings.TrimSpace(resource.Kind)
	if resource.Name == "" || resource.Kind == "" {
		return entity.RuntimePool{}, fmt.Errorf(
			"%w: runtime pool name and kind are required",
			entity.ErrInvalidInput,
		)
	}
	if hasNegative(resource.MaxTenants, resource.CurrentTenants) {
		return entity.RuntimePool{}, fmt.Errorf("%w: capacity values cannot be negative", entity.ErrInvalidInput)
	}
	return entity.RuntimePool{
		Name:           resource.Name,
		Kind:           resource.Kind,
		MaxTenants:     resource.MaxTenants,
		CurrentTenants: resource.CurrentTenants,
		Status:         activeStatus(resource.Status),
		Healthy:        resource.Healthy,
	}, nil
}

func toDatabaseClusterResource(cluster entity.DatabaseCluster) inputport.DatabaseClusterResource {
	return inputport.DatabaseClusterResource{
		Name:               cluster.Name,
		Engine:             cluster.Engine,
		Region:             cluster.Region,
		PlacementDB:        cluster.PlacementDB,
		MaxTenants:         cluster.MaxTenants,
		CurrentTenants:     cluster.CurrentTenants,
		MaxSchemas:         cluster.MaxSchemas,
		CurrentSchemas:     cluster.CurrentSchemas,
		MaxConnections:     cluster.MaxConnections,
		CurrentConnections: cluster.CurrentConnections,
		Status:             cluster.Status,
		Healthy:            cluster.Healthy,
		CreatedAt:          cluster.CreatedAt,
		UpdatedAt:          cluster.UpdatedAt,
	}
}

func toKubernetesClusterResource(cluster entity.KubernetesCluster) inputport.KubernetesClusterResource {
	namespaces := make([]inputport.KubernetesNamespaceResource, 0, len(cluster.Namespaces))
	for _, namespace := range cluster.Namespaces {
		namespaces = append(namespaces, inputport.KubernetesNamespaceResource{
			Name:           namespace.Name,
			MaxTenants:     namespace.MaxTenants,
			CurrentTenants: namespace.CurrentTenants,
			CPUMilli:       namespace.CPUMilli,
			MemoryMi:       namespace.MemoryMi,
			Status:         namespace.Status,
			Healthy:        namespace.Healthy,
		})
	}
	return inputport.KubernetesClusterResource{
		Name:       cluster.Name,
		Region:     cluster.Region,
		Namespaces: namespaces,
		Status:     cluster.Status,
		Healthy:    cluster.Healthy,
		CreatedAt:  cluster.CreatedAt,
		UpdatedAt:  cluster.UpdatedAt,
	}
}

func toRuntimePoolResource(pool entity.RuntimePool) inputport.RuntimePoolResource {
	return inputport.RuntimePoolResource{
		Name:           pool.Name,
		Kind:           pool.Kind,
		MaxTenants:     pool.MaxTenants,
		CurrentTenants: pool.CurrentTenants,
		Status:         pool.Status,
		Healthy:        pool.Healthy,
		CreatedAt:      pool.CreatedAt,
		UpdatedAt:      pool.UpdatedAt,
	}
}

func activeStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "active"
	}
	return strings.TrimSpace(status)
}

func hasNegative(values ...int) bool {
	for _, value := range values {
		if value < 0 {
			return true
		}
	}
	return false
}

func mapResourcePage[T any, R any](page collection.Page[T], mapper func(T) R) collection.Page[R] {
	items := make([]R, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, mapper(item))
	}
	return collection.Page[R]{
		Items:       items,
		Total:       page.Total,
		Page:        page.Page,
		PageSize:    page.PageSize,
		TotalPages:  page.TotalPages,
		HasNext:     page.HasNext,
		HasPrevious: page.HasPrevious,
	}
}
