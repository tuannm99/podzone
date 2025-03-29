package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuannm99/podzone/pkg/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Constants for context keys
const (
	UserIDKey        = "user_id"
	UserRoleKey      = "user_role"
	CorrelationIDKey = "correlation_id"
)

// GRPCAuthInterceptor creates a gRPC interceptor for authentication
func GRPCAuthInterceptor(authService *auth.Service, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Skip authentication for certain methods
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract token from metadata
		token, err := extractTokenFromGRPC(ctx)
		if err != nil {
			logger.Warn("Authentication failed: token extraction error", zap.Error(err))
			return nil, status.Errorf(codes.Unauthenticated, "missing or invalid auth token")
		}

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			logger.Warn("Authentication failed: token validation error", zap.Error(err))
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
		}

		// Add user info to context
		newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)
		newCtx = context.WithValue(newCtx, UserRoleKey, claims.Role)

		// Continue with the call
		return handler(newCtx, req)
	}
}

// HTTPAuthMiddleware creates an HTTP middleware for authentication
func HTTPAuthMiddleware(authService *auth.Service, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for public endpoints
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from header
		token, err := extractTokenFromHTTP(c.Request)
		if err != nil {
			logger.Warn("Authentication failed: token extraction error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid auth token"})
			return
		}

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			logger.Warn("Authentication failed: token validation error", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
			return
		}

		// Add user info to context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)

		c.Next()
	}
}

// CorrelationIDMiddleware adds a correlation ID to the context for request tracing
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID from header or generate a new one
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}

		// Add correlation ID to context
		c.Set(CorrelationIDKey, correlationID)

		// Add correlation ID to response header
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get correlation ID
		correlationID, _ := c.Get(CorrelationIDKey)

		// Log request
		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.String("correlation_id", correlationID.(string)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get correlation ID
				correlationID, _ := c.Get(CorrelationIDKey)

				// Log error
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("correlation_id", correlationID.(string)),
					zap.String("path", c.Request.URL.Path),
				)

				// Return error to client
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}()

		c.Next()
	}
}

// Helper functions

// extractTokenFromGRPC extracts the authorization token from gRPC metadata
func extractTokenFromGRPC(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	auth := values[0]
	prefix := "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", status.Errorf(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(auth, prefix), nil
}

// extractTokenFromHTTP extracts the authorization token from HTTP headers
func extractTokenFromHTTP(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	prefix := "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", fmt.Errorf("invalid authorization format")
	}

	return strings.TrimPrefix(auth, prefix), nil
}

// isPublicMethod checks if a gRPC method is public (doesn't require authentication)
func isPublicMethod(method string) bool {
	// Define your public methods here
	publicMethods := []string{
		"/catalog.ProductService/ListProducts",
		"/catalog.ProductService/GetProduct",
		"/catalog.ProductService/SearchProducts",
		"/user.UserService/Login",
		"/user.UserService/Register",
	}

	for _, m := range publicMethods {
		if m == method {
			return true
		}
	}

	return false
}

// isPublicPath checks if an HTTP path is public (doesn't require authentication)
func isPublicPath(path string) bool {
	// Define your public paths here
	publicPaths := []string{
		"/api/v1/products",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/health",
		"/metrics",
	}

	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

// generateCorrelationID generates a unique correlation ID
func generateCorrelationID() string {
	// Use a UUID library in a real implementation
	return fmt.Sprintf("cid-%d", time.Now().UnixNano())
}
