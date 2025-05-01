package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
)

func NewRedirectResponseModifier(logger *zap.Logger) runtime.ServeMuxOption {
	return runtime.WithForwardResponseOption(
		func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
			if loginResp, ok := resp.(*pb.GoogleLoginResponse); ok && loginResp.RedirectUrl != "" {
				logger.Info("Redirecting to OAuth provider", zap.String("url", loginResp.RedirectUrl))
				w.Header().Set("Location", loginResp.RedirectUrl)
				w.WriteHeader(http.StatusTemporaryRedirect)
				return nil
			}

			// if callbackResp, ok := resp.(*pb.GoogleCallbackResponse); ok && callbackResp.RedirectUrl != "" {
			// 	logger.Info("Redirecting to app after OAuth callback", zap.String("url", callbackResp.RedirectUrl))
			// 	w.Header().Set("Location", callbackResp.RedirectUrl)
			// 	w.WriteHeader(http.StatusTemporaryRedirect)
			// 	return nil
			// }

			return nil
		},
	)
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
