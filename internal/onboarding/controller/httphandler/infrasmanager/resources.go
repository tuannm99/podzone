package infrasmanager

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
	"github.com/tuannm99/podzone/pkg/collection"
)

type databaseClusterPage struct {
	Items    []inputport.DatabaseClusterResource `json:"items"`
	PageInfo collection.PageInfo                 `json:"pageInfo"`
}

type kubernetesClusterPage struct {
	Items    []inputport.KubernetesClusterResource `json:"items"`
	PageInfo collection.PageInfo                   `json:"pageInfo"`
}

type runtimePoolPage struct {
	Items    []inputport.RuntimePoolResource `json:"items"`
	PageInfo collection.PageInfo             `json:"pageInfo"`
}

func (c *Controller) ListDatabaseClusters(ctx *gin.Context) {
	query, ok := resourceCollectionQuery(ctx)
	if !ok {
		return
	}
	page, err := c.service.ListDatabaseClusters(ctx, query)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, databaseClusterPage{Items: page.Items, PageInfo: page.Info()})
}

func (c *Controller) UpsertDatabaseCluster(ctx *gin.Context) {
	var resource inputport.DatabaseClusterResource
	if !bindNamedResource(ctx, &resource, func() string { return resource.Name }) {
		return
	}
	if err := c.service.UpsertDatabaseCluster(ctx, resource); err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) CheckDatabaseClusterHealth(ctx *gin.Context) {
	resp, err := c.service.CheckDatabaseClusterHealth(ctx, ctx.Param("name"))
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *Controller) DeleteDatabaseCluster(ctx *gin.Context) {
	if err := c.service.DeleteDatabaseCluster(ctx, ctx.Param("name")); err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) ListKubernetesClusters(ctx *gin.Context) {
	query, ok := resourceCollectionQuery(ctx)
	if !ok {
		return
	}
	page, err := c.service.ListKubernetesClusters(ctx, query)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, kubernetesClusterPage{Items: page.Items, PageInfo: page.Info()})
}

func (c *Controller) UpsertKubernetesCluster(ctx *gin.Context) {
	var resource inputport.KubernetesClusterResource
	if !bindNamedResource(ctx, &resource, func() string { return resource.Name }) {
		return
	}
	if err := c.service.UpsertKubernetesCluster(ctx, resource); err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) DeleteKubernetesCluster(ctx *gin.Context) {
	if err := c.service.DeleteKubernetesCluster(ctx, ctx.Param("name")); err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) ListRuntimePools(ctx *gin.Context) {
	query, ok := resourceCollectionQuery(ctx)
	if !ok {
		return
	}
	page, err := c.service.ListRuntimePools(ctx, query)
	if err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, runtimePoolPage{Items: page.Items, PageInfo: page.Info()})
}

func (c *Controller) UpsertRuntimePool(ctx *gin.Context) {
	var resource inputport.RuntimePoolResource
	if !bindNamedResource(ctx, &resource, func() string { return resource.Name }) {
		return
	}
	if err := c.service.UpsertRuntimePool(ctx, resource); err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *Controller) DeleteRuntimePool(ctx *gin.Context) {
	if err := c.service.DeleteRuntimePool(ctx, ctx.Param("name")); err != nil {
		writeResourceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func resourceCollectionQuery(ctx *gin.Context) (collection.Query, bool) {
	query, err := collection.ParseURLValues(ctx.Request.URL.Query(), "collection.")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return collection.Query{}, false
	}
	return query, true
}

func bindNamedResource(ctx *gin.Context, target any, resourceName func() string) bool {
	if err := ctx.ShouldBindJSON(target); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return false
	}
	if resourceName() != ctx.Param("name") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "resource_name_mismatch",
			"message": "resource name must match the URL path",
		})
		return false
	}
	return true
}

func writeResourceError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrInvalidInput), errors.Is(err, collection.ErrInvalidQuery):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid_resource", "message": err.Error()})
	case errors.Is(err, entity.ErrResourceNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "resource_not_found", "message": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
	}
}
