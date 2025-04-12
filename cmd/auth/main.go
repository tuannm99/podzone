package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/pkg/config"
	"github.com/tuannm99/podzone/pkg/logging"
	"github.com/tuannm99/podzone/pkg/middleware"
	"github.com/tuannm99/podzone/services/auth"
)

func main() {
	config.LoadEnv()
	appEnv := os.Getenv("APP_ENV")
	logLevel := os.Getenv("DEFAULT_LOG_LEVEL")
	logger := logging.GetLoggerWithConfig(logLevel, appEnv)

	// Create auth server
	authServer, err := auth.NewAuthServer()
	if err != nil {
		logger.Fatal("Failed to create auth server", zap.Error(err))
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalChan)

	// Setup ports
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

	grpcGatewayMux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(auth.RedirectResponseModifier),
	)
	if err := pb.RegisterAuthServiceHandler(ctx, grpcGatewayMux, conn); err != nil {
		logger.Fatal("Failed to register service handler", zap.Error(err))
	}

	// API middleware
	middlewares := middleware.Default(logger)
	middlewares = append(middlewares, auth.AuthMiddleware(logger))
	handler := middleware.Chain(
		grpcGatewayMux,
		middlewares...,
	)

	// Start HTTP server
	gwServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: handler,
	}

	// Start the HTTP server in a goroutine
	go func() {
		logger.Info("gRPC-Gateway started",
			zap.String("address", "http://0.0.0.0:"+httpPort))
		if err := gwServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for termination signal
	<-signalChan
	logger.Info("Received shutdown signal, initiating graceful shutdown")

	// Create a deadline for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := gwServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	} else {
		logger.Info("HTTP server shutdown complete")
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()
	logger.Info("gRPC server shutdown complete")

	// Close the gRPC connection
	if err := conn.Close(); err != nil {
		logger.Error("Error closing gRPC connection", zap.Error(err))
	}

	authServer.Shutdown()

	logger.Info("Service shutdown complete")
}
