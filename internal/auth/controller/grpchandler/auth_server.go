package grpchandler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

type AuthServer struct {
	pbauthv1.UnimplementedAuthServiceServer
	authUC         inputport.AuthUsecase
	sessionRep     outputport.SessionRepository
	auditRep       outputport.AuditLogRepository
	userRepo       outputport.UserRepository
	jwtSecret      string
	jwtKey         string
	appRedirectURL string
}

func NewAuthServer(
	authUC inputport.AuthUsecase,
	sessionRep outputport.SessionRepository,
	auditRep outputport.AuditLogRepository,
	userRepo outputport.UserRepository,
	cfg config.AuthConfig,
) *AuthServer {
	return &AuthServer{
		authUC:         authUC,
		sessionRep:     sessionRep,
		auditRep:       auditRep,
		userRepo:       userRepo,
		jwtSecret:      cfg.JWTSecret,
		jwtKey:         cfg.JWTKey,
		appRedirectURL: cfg.AppRedirectURL,
	}
}

func (s *AuthServer) recordAudit(
	ctx context.Context,
	actorUserID uint,
	action string,
	resourceType string,
	resourceID string,
	tenantID string,
	payload map[string]any,
) {
	if s.auditRep == nil {
		return
	}
	payloadJSON := "{}"
	if len(payload) > 0 {
		if raw, err := json.Marshal(payload); err == nil {
			payloadJSON = string(raw)
		}
	}
	_ = s.auditRep.Create(ctx, entity.AuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  actorUserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TenantID:     tenantID,
		Status:       entity.AuditStatusSuccess,
		PayloadJSON:  payloadJSON,
		CreatedAt:    time.Now().UTC(),
	})
}

func authStatusError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, entity.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, entity.ErrWrongPassword):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, entity.ErrUserAlreadyExists),
		errors.Is(err, entity.ErrUsernameExisted),
		errors.Is(err, entity.ErrEmailExisted),
		errors.Is(err, entity.ErrInvalidSessionPolicy),
		errors.Is(err, entity.ErrInvalidUserID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, entity.ErrSessionNotFound),
		errors.Is(err, entity.ErrSessionRevoked),
		errors.Is(err, entity.ErrRefreshTokenInvalid),
		errors.Is(err, entity.ErrRefreshTokenExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *AuthServer) actorUserIDFromContext(ctx context.Context) (uint, error) {
	claims, err := s.claimsFromContext(ctx)
	if err != nil {
		return 0, err
	}
	if claims.UserID == 0 {
		return 0, errors.New("access token missing user_id")
	}
	return claims.UserID, nil
}

func (s *AuthServer) authorizedContext(ctx context.Context) (context.Context, uint, error) {
	userID, err := s.actorUserIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	return ctx, userID, nil
}

func (s *AuthServer) currentSessionFromAccessToken(accessToken string) (*entity.Session, error) {
	claims, err := s.claimsFromAccessToken(accessToken)
	if err != nil {
		return nil, err
	}
	return s.sessionRep.GetByID(context.Background(), claims.SessionID)
}

func (s *AuthServer) claimsFromContext(ctx context.Context) (*entity.JWTClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing request metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, errors.New("missing authorization header")
	}
	raw := strings.TrimSpace(values[0])
	if !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return nil, errors.New("invalid authorization header")
	}
	return s.claimsFromTokenString(strings.TrimSpace(raw[len("Bearer "):]))
}

func (s *AuthServer) claimsFromAccessToken(accessToken string) (*entity.JWTClaims, error) {
	tokenString := strings.TrimSpace(accessToken)
	if tokenString == "" {
		return nil, errors.New("missing access token")
	}
	return s.claimsFromTokenString(tokenString)
}

func (s *AuthServer) claimsFromTokenString(tokenString string) (*entity.JWTClaims, error) {
	claims := &entity.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	if s.jwtKey != "" && claims.Key != s.jwtKey {
		return nil, errors.New("invalid access token")
	}
	return claims, nil
}

func toUint(v uint64) (uint, error) {
	if v == 0 {
		return 0, entity.ErrInvalidUserID
	}
	return uint(v), nil
}
