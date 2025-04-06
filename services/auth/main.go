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
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/protobuf/proto"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
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

type StateStore struct {
	states map[string]time.Time
	mu     sync.Mutex
}

func NewStateStore() *StateStore {
	return &StateStore{
		states: make(map[string]time.Time),
	}
}

func (s *StateStore) Add(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = time.Now().Add(10 * time.Minute)
}

func (s *StateStore) Verify(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiry, exists := s.states[state]
	if !exists {
		return false
	}

	delete(s.states, state)
	return time.Now().Before(expiry)
}

func (s *StateStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for state, expiry := range s.states {
		if now.After(expiry) {
			delete(s.states, state)
		}
	}
}

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
	StateStore *StateStore
}

func (s *AuthServer) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleLoginResponse, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("error generating state: %v", err)
	}
	state := base64.StdEncoding.EncodeToString(b)

	s.StateStore.Add(state)

	url := googleOauthConfig.AuthCodeURL(state)

	return &pb.GoogleLoginResponse{
		RedirectUrl: url,
	}, nil
}

func (s *AuthServer) GoogleCallback(
	ctx context.Context,
	req *pb.GoogleCallbackRequest,
) (*pb.GoogleCallbackResponse, error) {
	if !s.StateStore.Verify(req.State) {
		return nil, fmt.Errorf("invalid state")
	}

	token, err := googleOauthConfig.Exchange(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange code: %v", err)
	}

	userInfo, err := getUserInfo(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("unable to get user info: %v", err)
	}

	jwtToken, err := createJWT(userInfo)
	if err != nil {
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

	return &pb.GoogleCallbackResponse{
		JwtToken:    jwtToken,
		RedirectUrl: redirectURL,
		UserInfo:    pbUserInfo,
	}, nil
}

func (s *AuthServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	if req.Token == "" {
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "empty token",
		}, nil
	}

	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "invalid token: " + err.Error(),
		}, nil
	}

	if !token.Valid {
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "invalid token",
		}, nil
	}

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
	return &pb.LogoutResponse{
		Success:     true,
		RedirectUrl: "/",
	}, nil
}

func getUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Error close body...")
		}
	}()

	var userInfo GoogleUserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func createJWT(userInfo *GoogleUserInfo) (string, error) {
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
		return "", err
	}

	return tokenString, nil
}

func RedirectResponseModifier(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	if loginResp, ok := resp.(*pb.GoogleLoginResponse); ok && loginResp.RedirectUrl != "" {
		w.Header().Set("Location", loginResp.RedirectUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return nil
	}

	if callbackResp, ok := resp.(*pb.GoogleCallbackResponse); ok && callbackResp.RedirectUrl != "" {
		w.Header().Set("Location", callbackResp.RedirectUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return nil
	}

	return nil
}

func AuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/auth/verify") && r.Method == "GET" {
			token := r.URL.Query().Get("token")

			if token == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if token != "" {
				jsonBody := fmt.Sprintf(`{"token": "%s"}`, token)
				r.Body = http.NoBody
				r.ContentLength = 0
				r.Header.Set("Content-Type", "application/json")
				r.Body = io.NopCloser(strings.NewReader(jsonBody))
				r.ContentLength = int64(len(jsonBody))
				r.Method = "POST"
			}
		}

		handler.ServeHTTP(w, r)
	})
}
