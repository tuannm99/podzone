package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go.uber.org/fx"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport"
	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
)

var (
	_ outputport.PlacementRouteReader = (*PlacementRouteReader)(nil)
	_ outputport.PlacementRouteWriter = (*PlacementRouteReader)(nil)
)

type PlacementRouteReader struct {
	kv  kvstores.KVStore
	cfg onboardingconfig.StoreProvisioningConfig
}

type PlacementRouteReaderParams struct {
	fx.In

	KV     kvstores.KVStore
	Config onboardingconfig.StoreProvisioningConfig
}

func NewPlacementRouteReader(params PlacementRouteReaderParams) *PlacementRouteReader {
	return &PlacementRouteReader{kv: params.KV, cfg: params.Config}
}

func (r *PlacementRouteReader) IsPlacementRouteReady(ctx context.Context, tenantID string) (bool, error) {
	key := "podzone/tenants/" + tenantID + "/placement"
	_, err := r.kv.Get(ctx, key)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, kvstores.ErrKeyNotFound) {
		return false, nil
	}
	return false, fmt.Errorf("read placement route %q: %w", key, err)
}

func (r *PlacementRouteReader) GetPlacementRoute(
	ctx context.Context,
	tenantID string,
) (*entity.PlacementRoute, error) {
	key := "podzone/tenants/" + tenantID + "/placement"
	raw, err := r.kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, kvstores.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("read placement route %q: %w", key, err)
	}

	var payload struct {
		ClusterName string `json:"cluster_name"`
		Mode        string `json:"mode"`
		DBName      string `json:"db_name"`
		SchemaName  string `json:"schema_name"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("decode placement route %q: %w", key, err)
	}
	return &entity.PlacementRoute{
		ClusterName: payload.ClusterName,
		Mode:        payload.Mode,
		DBName:      payload.DBName,
		SchemaName:  payload.SchemaName,
	}, nil
}

func (r *PlacementRouteReader) PublishPlacementRoute(
	ctx context.Context,
	tenantID string,
	allocation entity.PlacementAllocation,
) error {
	payload := map[string]string{
		"cluster_name": allocation.ClusterName,
		"mode":         allocation.Mode,
		"db_name":      allocation.DBName,
		"schema_name":  allocation.SchemaName,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal placement route: %w", err)
	}
	key := "podzone/tenants/" + tenantID + "/placement"
	if err := r.kv.Put(ctx, key, raw); err != nil {
		return fmt.Errorf("publish placement route %q: %w", key, err)
	}
	if err := r.publishClusterRoute(ctx, allocation.ClusterName); err != nil {
		return err
	}
	return nil
}

func (r *PlacementRouteReader) publishClusterRoute(ctx context.Context, clusterName string) error {
	if strings.TrimSpace(clusterName) == "" || strings.TrimSpace(r.cfg.AdminDSN) == "" {
		return nil
	}

	payload, err := clusterRoutePayload(r.cfg.AdminDSN)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal postgres cluster route: %w", err)
	}

	key := "podzone/postgres/clusters/" + clusterName
	if err := r.kv.Put(ctx, key, raw); err != nil {
		return fmt.Errorf("publish postgres cluster route %q: %w", key, err)
	}
	return nil
}

func clusterRoutePayload(dsn string) (map[string]interface{}, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres admin_dsn: %w", err)
	}
	port := 5432
	if rawPort := parsed.Port(); rawPort != "" {
		parsedPort, err := strconv.Atoi(rawPort)
		if err != nil {
			return nil, fmt.Errorf("parse postgres admin_dsn port: %w", err)
		}
		port = parsedPort
	}

	password, _ := parsed.User.Password()
	sslMode := parsed.Query().Get("sslmode")
	if sslMode == "" {
		sslMode = "disable"
	}
	return map[string]interface{}{
		"host":     parsed.Hostname(),
		"port":     port,
		"user":     parsed.User.Username(),
		"password": password,
		"ssl_mode": sslMode,
	}, nil
}
