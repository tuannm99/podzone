//go:build ignore

package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	authconfig "github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	authentity "github.com/tuannm99/podzone/internal/auth/domain/entity"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type cfg struct {
	TenantID    string
	TenantName  string
	TenantSlug  string
	Username    string
	Email       string
	Password    string
	FullName    string
	PGHost      string
	PGPort      string
	PGUser      string
	PGPassword  string
	PGSSLMode   string
	JWTSecret   string
	JWTKey      string
	OutputPath  string
	OwnerIDPath string
}

type seedOutput struct {
	TenantID     string `json:"tenantId"`
	TenantSlug   string `json:"tenantSlug"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	UserID       uint   `json:"userId"`
	SessionID    string `json:"sessionId"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	OutputPath   string `json:"outputPath,omitempty"`
}

func main() {
	cfg := loadCfg()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	authDB, err := openDB(cfg, "auth")
	if err != nil {
		fail("open auth db", err)
	}
	defer authDB.Close()

	iamDB, err := openDB(cfg, "iam")
	if err != nil {
		fail("open iam db", err)
	}
	defer iamDB.Close()

	out, err := seedAuth(ctx, authDB, iamDB, cfg)
	if err != nil {
		fail("seed auth bootstrap", err)
	}

	payload, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fail("marshal auth bootstrap output", err)
	}
	if cfg.OutputPath != "" {
		if err := os.WriteFile(cfg.OutputPath, payload, 0o600); err != nil {
			fail("write auth bootstrap output", err)
		}
	}
	if cfg.OwnerIDPath != "" {
		if err := os.WriteFile(cfg.OwnerIDPath, []byte(fmt.Sprint(out.UserID)), 0o600); err != nil {
			fail("write auth owner ID output", err)
		}
	}

	fmt.Println(string(payload))
	fmt.Println("")
	fmt.Println("Dev auth bootstrap ready.")
	fmt.Printf("Login: %s / %s\n", cfg.Username, cfg.Password)
	fmt.Printf("Tenant: %s\n", cfg.TenantID)
	if cfg.OutputPath != "" {
		fmt.Printf("Saved token bundle: %s\n", cfg.OutputPath)
	}
}

func loadCfg() cfg {
	tenantID := envOr("TENANT_ID", "tenant-dev")
	username := envOr("DEV_USERNAME", "devowner")
	return cfg{
		TenantID:    tenantID,
		TenantName:  envOr("TENANT_NAME", "Demo POD Tenant"),
		TenantSlug:  envOr("TENANT_SLUG", sanitizeSlug(tenantID)),
		Username:    username,
		Email:       envOr("DEV_EMAIL", username+"@podzone.dev"),
		Password:    envOr("DEV_PASSWORD", "DevPass123!"),
		FullName:    envOr("DEV_FULL_NAME", "Dev Owner"),
		PGHost:      envOr("PG_HOST", "localhost"),
		PGPort:      envOr("PG_PORT", "5432"),
		PGUser:      envOr("PG_USER", "postgres"),
		PGPassword:  envOr("PG_PASSWORD", "postgres"),
		PGSSLMode:   envOr("PG_SSL_MODE", "disable"),
		JWTSecret:   envOr("JWT_SECRET", "dev-secret"),
		JWTKey:      envOr("JWT_KEY", ""),
		OutputPath:  envOr("AUTH_BOOTSTRAP_OUTPUT", "/tmp/podzone-dev-auth.json"),
		OwnerIDPath: strings.TrimSpace(os.Getenv("DEV_OWNER_ID_OUTPUT")),
	}
}

func seedAuth(ctx context.Context, authDB *sqlx.DB, iamDB *sqlx.DB, cfg cfg) (*seedOutput, error) {
	authTx, err := authDB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = authTx.Rollback() }()

	now := time.Now().UTC()
	userID, err := upsertUser(ctx, authTx, cfg, now)
	if err != nil {
		return nil, err
	}
	if err := authTx.Commit(); err != nil {
		return nil, err
	}

	iamTx, err := iamDB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = iamTx.Rollback() }()

	if err := upsertTenant(ctx, iamTx, cfg, now); err != nil {
		return nil, err
	}

	roleID, err := getRoleID(ctx, iamTx, iamdomain.RoleTenantOwner)
	if err != nil {
		return nil, err
	}
	if err := upsertTenantMembership(ctx, iamTx, cfg.TenantID, userID, roleID, now); err != nil {
		return nil, err
	}

	platformRoleID, err := getRoleID(ctx, iamTx, iamdomain.RolePlatformOwner)
	if err != nil {
		return nil, fmt.Errorf("load platform owner role: %w", err)
	}
	if err := upsertPlatformRole(ctx, iamTx, userID, platformRoleID, now); err != nil {
		return nil, err
	}
	if err := iamTx.Commit(); err != nil {
		return nil, err
	}

	sessionID := uuid.NewString()
	refreshTokenRaw, refreshTokenHash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	sessionTx, err := authDB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sessionTx.Rollback() }()
	if err := revokeExistingSessions(ctx, sessionTx, userID, now); err != nil {
		return nil, err
	}
	if err := insertSession(ctx, sessionTx, sessionID, userID, cfg.TenantID, now); err != nil {
		return nil, err
	}
	if err := insertRefreshToken(ctx, sessionTx, sessionID, refreshTokenHash, now); err != nil {
		return nil, err
	}

	user := authentity.User{
		Id:       userID,
		Username: cfg.Username,
		Email:    cfg.Email,
		FullName: cfg.FullName,
	}
	tokenUC := authdomain.NewTokenUsecase(authconfig.AuthConfig{
		JWTSecret: cfg.JWTSecret,
		JWTKey:    cfg.JWTKey,
	})
	accessToken, err := tokenUC.CreateJwtTokenForSession(user, cfg.TenantID, sessionID)
	if err != nil {
		return nil, err
	}

	if err := sessionTx.Commit(); err != nil {
		return nil, err
	}

	return &seedOutput{
		TenantID:     cfg.TenantID,
		TenantSlug:   cfg.TenantSlug,
		Username:     cfg.Username,
		Email:        cfg.Email,
		Password:     cfg.Password,
		UserID:       userID,
		SessionID:    sessionID,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenRaw,
		OutputPath:   cfg.OutputPath,
	}, nil
}

func openDB(cfg cfg, dbName string) (*sqlx.DB, error) {
	dsn := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.PGUser, cfg.PGPassword),
		Host:   fmt.Sprintf("%s:%s", cfg.PGHost, cfg.PGPort),
		Path:   "/" + dbName,
		RawQuery: url.Values{
			"sslmode": []string{cfg.PGSSLMode},
		}.Encode(),
	}).String()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func upsertUser(ctx context.Context, tx *sqlx.Tx, cfg cfg, now time.Time) (uint, error) {
	hashed, err := authentity.GeneratePasswordHash(cfg.Password)
	if err != nil {
		return 0, err
	}

	var userID uint
	err = tx.GetContext(ctx, &userID, `
		INSERT INTO users (
			username, email, password, full_name, initial_from, dob, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (email) DO UPDATE SET
			username = EXCLUDED.username,
			password = EXCLUDED.password,
			full_name = EXCLUDED.full_name,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`, cfg.Username, cfg.Email, hashed, cfg.FullName, "podzone-dev-seed", now, now, now)
	return userID, err
}

func upsertTenant(ctx context.Context, tx *sqlx.Tx, cfg cfg, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO tenants (id, slug, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug,
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`, cfg.TenantID, cfg.TenantSlug, cfg.TenantName, now, now)
	return err
}

func getRoleID(ctx context.Context, tx *sqlx.Tx, roleName string) (uint64, error) {
	var roleID uint64
	err := tx.GetContext(ctx, &roleID, `SELECT id FROM iam_roles WHERE name = $1`, roleName)
	return roleID, err
}

func upsertTenantMembership(
	ctx context.Context,
	tx *sqlx.Tx,
	tenantID string,
	userID uint,
	roleID uint64,
	now time.Time,
) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO tenant_memberships (tenant_id, user_id, role_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', $4, $5)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET
			role_id = EXCLUDED.role_id,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, tenantID, userID, roleID, now, now)
	return err
}

func upsertPlatformRole(ctx context.Context, tx *sqlx.Tx, userID uint, roleID uint64, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO user_platform_roles (user_id, role_id, status, created_at, updated_at)
		VALUES ($1, $2, 'active', $3, $4)
		ON CONFLICT (user_id, role_id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, userID, roleID, now, now)
	return err
}

func revokeExistingSessions(ctx context.Context, tx *sqlx.Tx, userID uint, now time.Time) error {
	if _, err := tx.ExecContext(ctx, `
		UPDATE auth_sessions
		SET status = 'revoked', revoked_at = $2, updated_at = $2
		WHERE user_id = $1 AND status = 'active'
	`, userID, now); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE auth_refresh_tokens
		SET revoked_at = $2, updated_at = $2
		WHERE session_id IN (
			SELECT id FROM auth_sessions WHERE user_id = $1
		) AND revoked_at IS NULL
	`, userID, now)
	return err
}

func insertSession(
	ctx context.Context,
	tx *sqlx.Tx,
	sessionID string,
	userID uint,
	tenantID string,
	now time.Time,
) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO auth_sessions (
			id, user_id, active_tenant_id, status, created_at, updated_at, expires_at, revoked_at
		) VALUES ($1, $2, $3, 'active', $4, $5, $6, NULL)
	`, sessionID, userID, tenantID, now, now, now.Add(30*24*time.Hour))
	return err
}

func insertRefreshToken(ctx context.Context, tx *sqlx.Tx, sessionID, tokenHash string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO auth_refresh_tokens (
			id, session_id, token_hash, expires_at, created_at, updated_at, revoked_at, replaced_by_token_id
		) VALUES ($1, $2, $3, $4, $5, $6, NULL, NULL)
	`, uuid.NewString(), sessionID, tokenHash, now.Add(30*24*time.Hour), now, now)
	return err
}

func newRefreshToken() (raw string, hash string, err error) {
	buf := make([]byte, 48)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, authentity.HashToken(raw), nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func sanitizeSlug(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, "_", "-")
	v = strings.ReplaceAll(v, " ", "-")
	if v == "" {
		return "tenant-dev"
	}
	return v
}

func fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", step, err)
	os.Exit(1)
}
