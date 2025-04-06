package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/services/auth"
)

func main() {
	if os.Getenv("GOOGLE_CLIENT_ID") == "" || os.Getenv("GOOGLE_CLIENT_SECRET") == "" {
		log.Fatal("Environment variables GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}
	if os.Getenv("OAUTH_REDIRECT_URL") == "" {
		log.Fatal("Environment variable OAUTH_REDIRECT_URL must be set")
	}
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("Environment variable JWT_SECRET must be set")
	}
	if os.Getenv("APP_REDIRECT_URL") == "" {
		log.Fatal("Environment variable APP_REDIRECT_URL must be set")
	}

	stateStore := auth.NewStateStore()

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
		log.Fatalf("failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	authServer := &auth.AuthServer{
		StateStore: stateStore,
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	log.Printf("gRPC auth server running on port %s", grpcPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
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
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	grpcGatewayMux := runtime.NewServeMux(
		runtime.WithForwardResponseOption(auth.RedirectResponseModifier),
	)

	err = pb.RegisterAuthServiceHandler(ctx, grpcGatewayMux, conn)
	if err != nil {
		log.Fatalf("Failed to register service handler: %v", err)
	}

	handler := auth.AuthMiddleware(grpcGatewayMux)

	gwServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: handler,
	}

	log.Printf("gRPC-Gateway running at http://0.0.0.0:%s", httpPort)
	log.Fatalln(gwServer.ListenAndServe())
}
