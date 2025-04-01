package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
	stateStore *StateStore
}

func (s *AuthServer) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleLoginResponse, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("lỗi tạo state: %v", err)
	}
	state := base64.StdEncoding.EncodeToString(b)

	s.stateStore.Add(state)

	url := googleOauthConfig.AuthCodeURL(state)

	return &pb.GoogleLoginResponse{
		RedirectUrl: url,
	}, nil
}

func (s *AuthServer) GoogleCallback(
	ctx context.Context,
	req *pb.GoogleCallbackRequest,
) (*pb.GoogleCallbackResponse, error) {
	if !s.stateStore.Verify(req.State) {
		return nil, fmt.Errorf("state không hợp lệ")
	}

	token, err := googleOauthConfig.Exchange(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("không thể trao đổi code: %v", err)
	}

	userInfo, err := getUserInfo(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("không thể lấy thông tin người dùng: %v", err)
	}

	jwtToken, err := createJWT(userInfo)
	if err != nil {
		return nil, fmt.Errorf("lỗi tạo JWT: %v", err)
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
			Error:   "token trống",
		}, nil
	}

	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "token không hợp lệ: " + err.Error(),
		}, nil
	}

	if !token.Valid {
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "token không hợp lệ",
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
	defer resp.Body.Close()

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

func ginGoogleLoginHandler(stateStore *StateStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi tạo state"})
			return
		}
		state := base64.StdEncoding.EncodeToString(b)

		stateStore.Add(state)

		redirectAfterLogin := c.Query("redirect_after_login")
		if redirectAfterLogin != "" {
			// Lưu redirect_after_login vào state (có thể lưu vào cache/database)
		}

		url := googleOauthConfig.AuthCodeURL(state)

		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func ginGoogleCallbackHandler(stateStore *StateStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := c.Query("state")
		code := c.Query("code")

		if !stateStore.Verify(state) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "State không hợp lệ"})
			return
		}

		token, err := googleOauthConfig.Exchange(c, code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể trao đổi code: " + err.Error()})
			return
		}

		userInfo, err := getUserInfo(token.AccessToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể lấy thông tin người dùng: " + err.Error()})
			return
		}

		jwtToken, err := createJWT(userInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi tạo JWT: " + err.Error()})
			return
		}

		redirectURL := fmt.Sprintf("%s?token=%s", os.Getenv("APP_REDIRECT_URL"), jwtToken)

		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}
}

func ginVerifyTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		token = c.Query("token")

		if token == "" {
			var req struct {
				Token string `json:"token"`
			}
			if err := c.ShouldBindJSON(&req); err == nil {
				token = req.Token
			}
		}

		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"is_valid": false,
				"error":    "Token trống",
			})
			return
		}

		claims := &JWTClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"is_valid": false,
				"error":    "Token không hợp lệ: " + err.Error(),
			})
			return
		}

		if !parsedToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"is_valid": false,
				"error":    "Token không hợp lệ",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"is_valid": true,
			"user_info": gin.H{
				"id":    claims.Sub,
				"email": claims.Email,
				"name":  claims.Name,
			},
		})
	}
}

func ginLogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"redirect_url": "/",
		})
	}
}

func main() {
	if os.Getenv("GOOGLE_CLIENT_ID") == "" || os.Getenv("GOOGLE_CLIENT_SECRET") == "" {
		log.Fatal("Cần thiết lập biến môi trường GOOGLE_CLIENT_ID và GOOGLE_CLIENT_SECRET")
	}
	if os.Getenv("OAUTH_REDIRECT_URL") == "" {
		log.Fatal("Cần thiết lập biến môi trường OAUTH_REDIRECT_URL")
	}
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("Cần thiết lập biến môi trường JWT_SECRET")
	}
	if os.Getenv("APP_REDIRECT_URL") == "" {
		log.Fatal("Cần thiết lập biến môi trường APP_REDIRECT_URL")
	}

	stateStore := NewStateStore()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			stateStore.Cleanup()
		}
	}()

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("không thể lắng nghe cổng gRPC %s: %v", grpcPort, err)
	}

	authServer := &AuthServer{
		stateStore: stateStore,
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	log.Printf("gRPC auth server đang chạy tại cổng %s", grpcPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("không thể phục vụ gRPC: %v", err)
		}
	}()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		"localhost:"+grpcPort,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Không thể kết nối đến gRPC server: %v", err)
	}
	defer conn.Close()

	grpcGatewayMux := runtime.NewServeMux()
	err = pb.RegisterAuthServiceHandler(ctx, grpcGatewayMux, conn)
	if err != nil {
		log.Fatalf("Không thể đăng ký service handler: %v", err)
	}

	router := gin.Default()

	router.GET("/v1/auth/google/login", ginGoogleLoginHandler(stateStore))
	router.GET("/v1/auth/google/callback", ginGoogleCallbackHandler(stateStore))
	router.GET("/v1/auth/verify", ginVerifyTokenHandler())
	router.POST("/v1/auth/verify", ginVerifyTokenHandler())
	router.GET("/v1/auth/logout", ginLogoutHandler())

	router.NoRoute(func(c *gin.Context) {
		grpcGatewayMux.ServeHTTP(c.Writer, c.Request)
	})

	log.Printf("HTTP Auth server đang chạy tại cổng %s", httpPort)
	if err := router.Run(":" + httpPort); err != nil {
		log.Fatalf("không thể khởi động HTTP server: %v", err)
	}
}
