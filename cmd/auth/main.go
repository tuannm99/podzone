package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/pkg/logging"
	"github.com/tuannm99/podzone/services/auth"
)

func main() {
	// Determine environment
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	// Initialize logger
	logger, err := logging.NewLogger("info", env)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Check required environment variables
	if os.Getenv("GOOGLE_CLIENT_ID") == "" || os.Getenv("GOOGLE_CLIENT_SECRET") == "" {
		logger.Fatal("Environment variables GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}
	if os.Getenv("OAUTH_REDIRECT_URL") == "" {
		logger.Fatal("Environment variable OAUTH_REDIRECT_URL must be set")
	}
	if os.Getenv("JWT_SECRET") == "" {
		logger.Fatal("Environment variable JWT_SECRET must be set")
	}
	if os.Getenv("APP_REDIRECT_URL") == "" {
		logger.Fatal("Environment variable APP_REDIRECT_URL must be set")
	}

	// Initialize auth server with logger
	authServer, err := auth.NewAuthServer(env)
	if err != nil {
		logger.Fatal("Failed to create auth server", zap.Error(err))
	}

	// Start state cleanup in background
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	authServer.StartCleanupRoutine(ctx)

	// Get port configuration
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatal("Failed to listen on gRPC port",
			zap.String("port", grpcPort),
			zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	logger.Info("gRPC auth server started", zap.String("port", grpcPort))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Set up gRPC Gateway
	conn, err := grpc.NewClient(
		"localhost:"+grpcPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error("Error closing gRPC connection", zap.Error(err))
		}
	}()

	grpcGatewayMux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(auth.RedirectResponseModifier),
	)

	if err := pb.RegisterAuthServiceHandler(ctx, grpcGatewayMux, conn); err != nil {
		logger.Fatal("Failed to register service handler", zap.Error(err))
	}

	// Apply combined middleware (auth + logging)
	handler := auth.CombinedMiddleware(logger, grpcGatewayMux)

	// Start HTTP server
	gwServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: handler,
	}

	logger.Info("gRPC-Gateway started",
		zap.String("address", "http://0.0.0.0:"+httpPort))

	if err := gwServer.ListenAndServe(); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

