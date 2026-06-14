package grpchandler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/tuannm99/podzone/internal/auth/config"
	authdomain "github.com/tuannm99/podzone/internal/auth/domain"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	inputmocks "github.com/tuannm99/podzone/internal/auth/domain/inputport/mocks"
	outputmocks "github.com/tuannm99/podzone/internal/auth/domain/outputport/mocks"
)

var testAuthCfg = config.AuthConfig{
	JWTSecret:      "secret",
	JWTKey:         "app-key",
	AppRedirectURL: "https://app.example.com/auth/google/callback",
}

func newAuthServer(t *testing.T) (
	*AuthServer,
	*inputmocks.MockAuthUsecase,
	*outputmocks.MockSessionRepository,
	*outputmocks.MockAuditLogRepository,
	*outputmocks.MockUserRepository,
) {
	t.Helper()
	authUC := inputmocks.NewMockAuthUsecase(t)
	sessionRepo := outputmocks.NewMockSessionRepository(t)
	auditRepo := outputmocks.NewMockAuditLogRepository(t)
	userRepo := outputmocks.NewMockUserRepository(t)
	return NewAuthServer(
		authUC,
		sessionRepo,
		auditRepo,
		userRepo,
		testAuthCfg,
	), authUC, sessionRepo, auditRepo, userRepo
}

func authContextForUser(t *testing.T, userID uint) context.Context {
	t.Helper()
	token, err := authdomain.NewTokenUsecase(testAuthCfg).
		CreateJwtTokenForSession(entity.User{Id: userID}, "", "session-test")
	require.NoError(t, err)
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}

func accessTokenForSession(t *testing.T, user entity.User, activeTenantID, sessionID string) string {
	t.Helper()
	token, err := authdomain.NewTokenUsecase(testAuthCfg).
		CreateJwtTokenForSession(user, activeTenantID, sessionID)
	require.NoError(t, err)
	return token
}

func sessionWithID(id string, userID uint) *entity.Session {
	return &entity.Session{
		ID:        id,
		UserID:    userID,
		Status:    entity.SessionStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().Add(time.Hour).UTC(),
	}
}

func expectAuditMaybe(auditRepo *outputmocks.MockAuditLogRepository) {
	auditRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Maybe()
}
