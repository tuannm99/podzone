package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/protobuf/proto"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/pkg/logging"
	"github.com/tuannm99/podzone/pkg/persistents/redis"
)

var (
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
)

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
}

type JWTClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
	jwt.StandardClaims
}

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	StateStore  *RedisStateStore
	logger      *zap.Logger
	redisClient *redis.Client
}

// NewAuthServer creates a new auth server with logging and Redis state store
func NewAuthServer() (*AuthServer, error) {
	logger := logging.GetLogger()

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisConfig := redis.DefaultConfig()
	if redisAddr != "" {
		redisConfig.Addr = redisAddr
	}
	if redisPassword != "" {
		redisConfig.Password = redisPassword
	}

	redisClient, err := redis.NewClient(redisConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %v", err)
	}

	stateStore := NewRedisStateStore(redisClient, logger)

	return &AuthServer{
		StateStore:  stateStore,
		logger:      logger,
		redisClient: redisClient,
	}, nil
}

func (s *AuthServer) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleLoginResponse, error) {
	s.logger.Info("Google login request received")

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		s.logger.Error("Error generating state", zap.Error(err))
		return nil, fmt.Errorf("error generating state: %v", err)
	}
	state := base64.StdEncoding.EncodeToString(b)

	if err := s.StateStore.Add(state); err != nil {
		return nil, fmt.Errorf("error storing state: %v", err)
	}

	url := googleOauthConfig.AuthCodeURL(state)
	s.logger.Info("Generated OAuth URL", zap.String("redirect_url", url))

	return &pb.GoogleLoginResponse{
		RedirectUrl: url,
	}, nil
}

func (s *AuthServer) GoogleCallback(
	ctx context.Context,
	req *pb.GoogleCallbackRequest,
) (*pb.GoogleCallbackResponse, error) {
	s.logger.Info("Google callback request received",
		zap.String("state", req.State),
		zap.Bool("has_code", req.Code != ""))

	if !s.StateStore.Verify(req.State) {
		s.logger.Warn("Invalid state received", zap.String("state", req.State))
		return nil, fmt.Errorf("invalid state")
	}

	token, err := googleOauthConfig.Exchange(ctx, req.Code)
	if err != nil {
		s.logger.Error("Unable to exchange code", zap.Error(err))
		return nil, fmt.Errorf("unable to exchange code: %v", err)
	}

	userInfo, err := s.getUserInfo(token.AccessToken)
	if err != nil {
		s.logger.Error("Unable to get user info", zap.Error(err))
		return nil, fmt.Errorf("unable to get user info: %v", err)
	}

	s.logger.Info("User authenticated successfully",
		zap.String("email", userInfo.Email),
		zap.String("user_id", userInfo.Sub))

	jwtToken, err := s.createJWT(userInfo)
	if err != nil {
		s.logger.Error("Error creating JWT", zap.Error(err))
		return nil, fmt.Errorf("error creating JWT: %v", err)
	}

	pbUserInfo := &pb.UserInfo{
		Id:            userInfo.Sub,
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		GivenName:     userInfo.GivenName,
		FamilyName:    userInfo.FamilyName,
		Picture:       userInfo.Picture,
		EmailVerified: userInfo.EmailVerified,
	}

	redirectURL := fmt.Sprintf("%s?token=%s", os.Getenv("APP_REDIRECT_URL"), jwtToken)
	s.logger.Info("Redirecting user to app", zap.String("redirect_url", redirectURL))

	return &pb.GoogleCallbackResponse{
		JwtToken:    jwtToken,
		RedirectUrl: redirectURL,
		UserInfo:    pbUserInfo,
	}, nil
}

func (s *AuthServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	s.logger.Debug("Token verification request received")

	if req.Token == "" {
		s.logger.Warn("Empty token received")
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "empty token",
		}, nil
	}

	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (any, error) {
		return jwtSecret, nil
	})
	if err != nil {
		s.logger.Warn("Invalid token", zap.Error(err))
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "invalid token: " + err.Error(),
		}, nil
	}

	if !token.Valid {
		s.logger.Warn("Token validation failed")
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "invalid token",
		}, nil
	}

	s.logger.Info("Token verified successfully",
		zap.String("user_id", claims.Sub),
		zap.String("email", claims.Email))

	return &pb.VerifyTokenResponse{
		IsValid: true,
		UserInfo: &pb.UserInfo{
			Id:    claims.Sub,
			Email: claims.Email,
			Name:  claims.Name,
		},
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	s.logger.Info("Logout request received")

	return &pb.LogoutResponse{
		Success:     true,
		RedirectUrl: "/",
	}, nil
}

func (s *AuthServer) getUserInfo(accessToken string) (*GoogleUserInfo, error) {
	s.logger.Debug("Fetching user info from Google")
	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + accessToken)
	if err != nil {
		s.logger.Error("Failed to get user info from Google", zap.Error(err))
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			s.logger.Warn("Error closing response body", zap.Error(err))
		}
	}()

	var userInfo GoogleUserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		s.logger.Error("Failed to decode user info", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Successfully retrieved user info",
		zap.String("email", userInfo.Email),
		zap.String("id", userInfo.Sub))
	return &userInfo, nil
}

func (s *AuthServer) createJWT(userInfo *GoogleUserInfo) (string, error) {
	s.logger.Debug("Creating JWT token", zap.String("user_id", userInfo.Sub))

	claims := JWTClaims{
		Email: userInfo.Email,
		Name:  userInfo.Name,
		Sub:   userInfo.Sub,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		s.logger.Error("Failed to sign JWT token", zap.Error(err))
		return "", err
	}

	s.logger.Debug("JWT token created successfully")
	return tokenString, nil
}

func RedirectResponseModifier(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	logger := logging.GetLogger()

	if loginResp, ok := resp.(*pb.GoogleLoginResponse); ok && loginResp.RedirectUrl != "" {
		logger.Info("Redirecting to OAuth provider", zap.String("url", loginResp.RedirectUrl))
		w.Header().Set("Location", loginResp.RedirectUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return nil
	}

	if callbackResp, ok := resp.(*pb.GoogleCallbackResponse); ok && callbackResp.RedirectUrl != "" {
		logger.Info("Redirecting to app after OAuth callback", zap.String("url", callbackResp.RedirectUrl))
		w.Header().Set("Location", callbackResp.RedirectUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return nil
	}

	return nil
}

func AuthMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/auth/v1/verify") && r.Method == "GET" {
				token := r.URL.Query().Get("token")

				if token == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						token = strings.TrimPrefix(authHeader, "Bearer ")
						logger.Debug("Extracted token from Authorization header")
					}
				} else {
					logger.Debug("Extracted token from query parameter")
				}

				if token != "" {
					logger.Debug("Converting GET request to POST for token verification")
					jsonBody := fmt.Sprintf(`{"token": "%s"}`, token)
					r.Body = http.NoBody
					r.ContentLength = 0
					r.Header.Set("Content-Type", "application/json")
					r.Body = io.NopCloser(strings.NewReader(jsonBody))
					r.ContentLength = int64(len(jsonBody))
					r.Method = "POST"
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Shutdown gracefully closes resources
func (s *AuthServer) Shutdown() {
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			s.logger.Error("Error closing Redis connection", zap.Error(err))
		}
	}
	s.logger.Info("Auth server resources cleaned up")
}
