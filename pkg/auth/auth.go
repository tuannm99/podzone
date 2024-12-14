package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

// TokenType defines the type of token
type TokenType string

const (
	// AccessToken is a short-lived token for API access
	AccessToken TokenType = "access"
	// RefreshToken is a long-lived token for obtaining new access tokens
	RefreshToken TokenType = "refresh"
)

// UserRole defines the user's role
type UserRole string

const (
	// AdminRole has full access to the system
	AdminRole UserRole = "admin"
	// CustomerRole has access to customer-related functionality
	CustomerRole UserRole = "customer"
	// EmployeeRole has access to employee-related functionality
	EmployeeRole UserRole = "employee"
)

// Claims represents JWT claims
type Claims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	TokenID   string    `json:"token_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// User represents user data
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository defines methods for user data access
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// TokenRepository defines methods for token management
type TokenRepository interface {
	Store(ctx context.Context, tokenID string, userID string, expiresAt time.Time) error
	Validate(ctx context.Context, tokenID string) (string, error)
	Revoke(ctx context.Context, tokenID string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

// Config holds service configuration
type Config struct {
	AccessTokenSecret          string        `mapstructure:"access_token_secret"`
	RefreshTokenSecret         string        `mapstructure:"refresh_token_secret"`
	AccessTokenExpiryDuration  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiryDuration time.Duration `mapstructure:"refresh_token_expiry"`
}

// Service provides authentication and authorization functionality
type Service struct {
	config    Config
	userRepo  UserRepository
	tokenRepo TokenRepository
}

// NewService creates a new authentication service
func NewService(config Config, userRepo UserRepository, tokenRepo TokenRepository) *Service {
	return &Service{
		config:    config,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

// Register creates a new user
func (s *Service) Register(ctx context.Context, user *User) error {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return err
	}
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Set user role to customer if not specified
	if user.Role == "" {
		user.Role = CustomerRole
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Create user
	return s.userRepo.Create(ctx, user)
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, email, password string) (string, string, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.GenerateAccessToken(ctx, user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.GenerateRefreshToken(ctx, user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessToken generates a new access token
func (s *Service) GenerateAccessToken(ctx context.Context, user *User) (string, error) {
	return s.generateToken(ctx, user, AccessToken, s.config.AccessTokenSecret, s.config.AccessTokenExpiryDuration)
}

// GenerateRefreshToken generates a new refresh token
func (s *Service) GenerateRefreshToken(ctx context.Context, user *User) (string, error) {
	return s.generateToken(ctx, user, RefreshToken, s.config.RefreshTokenSecret, s.config.RefreshTokenExpiryDuration)
}

// generateToken creates a new token
func (s *Service) generateToken(
	ctx context.Context,
	user *User,
	tokenType TokenType,
	secret string,
	expiry time.Duration,
) (string, error) {
	// Generate unique token ID
	tokenID := generateUUID()
	expiresAt := time.Now().Add(expiry)

	// Create claims
	claims := Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		TokenID:   tokenID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	// Store token in repository (for blacklisting/validation)
	if err := s.tokenRepo.Store(ctx, tokenID, user.ID, expiresAt); err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a token and returns its claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	// Parse token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Use correct secret based on token type
		if claims.TokenType == AccessToken {
			return []byte(s.config.AccessTokenSecret), nil
		}
		return []byte(s.config.RefreshTokenSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshTokens validates a refresh token and issues new tokens
func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	// Verify token type
	if claims.TokenType != RefreshToken {
		return "", "", ErrInvalidToken
	}

	// Check if token is valid in repository
	userID, err := s.tokenRepo.Validate(ctx, claims.TokenID)
	if err != nil {
		return "", "", err
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	// Revoke old refresh token
	if err := s.tokenRepo.Revoke(ctx, claims.TokenID); err != nil {
		return "", "", err
	}

	// Generate new tokens
	newAccessToken, err := s.GenerateAccessToken(ctx, user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.GenerateRefreshToken(ctx, user)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout revokes a refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return err
	}

	// Verify token type
	if claims.TokenType != RefreshToken {
		return ErrInvalidToken
	}

	// Revoke token
	return s.tokenRepo.Revoke(ctx, claims.TokenID)
}

// LogoutAll revokes all refresh tokens for a user
func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	return s.tokenRepo.RevokeAllForUser(ctx, userID)
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(ctx context.Context, userID string, oldPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	// Save user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Revoke all tokens
	return s.tokenRepo.RevokeAllForUser(ctx, userID)
}

// RedisTokenRepository implements TokenRepository using Redis
type RedisTokenRepository struct {
	client *redis.Client
}

// NewRedisTokenRepository creates a new Redis token repository
func NewRedisTokenRepository(client *redis.Client) *RedisTokenRepository {
	return &RedisTokenRepository{
		client: client,
	}
}

// Store stores a token in Redis
func (r *RedisTokenRepository) Store(ctx context.Context, tokenID string, userID string, expiresAt time.Time) error {
	// Calculate TTL
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("token already expired")
	}

	// Store token
	key := fmt.Sprintf("token:%s", tokenID)
	return r.client.Set(ctx, key, userID, ttl).Err()
}

// Validate checks if a token is valid
func (r *RedisTokenRepository) Validate(ctx context.Context, tokenID string) (string, error) {
	key := fmt.Sprintf("token:%s", tokenID)
	userID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrInvalidToken
		}
		return "", err
	}
	return userID, nil
}

// Revoke invalidates a token
func (r *RedisTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("token:%s", tokenID)
	return r.client.Del(ctx, key).Err()
}

// RevokeAllForUser revokes all tokens for a user
func (r *RedisTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	// In a real implementation, you would need to maintain a set of tokens per user
	// For simplicity, this implementation only shows the interface
	return nil
}

// Helper function to generate UUID for token IDs
func generateUUID() string {
	// In a real implementation, use a proper UUID library
	return fmt.Sprintf("tok-%d", time.Now().UnixNano())
}
