package main

import (
	c "calendar"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	protobufs "scheduling"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "healthy")
}

// Handle simple kubernetes or other container orchestration health checks
func healthCheck() {
	healthCheckHost := os.Getenv("HEALTH_CHECK_HOST")

	if healthCheckHost == "" {
		healthCheckHost = ":8080"
		log.Printf("Defaulting to host health check at %s", healthCheckHost)
	}

	http.HandleFunc("/healthz", handler)
	log.Printf("Health check listening at %s", healthCheckHost)

	// If we failed to bind a health check, panic to alert operators
	log.Fatal(http.ListenAndServe(healthCheckHost, nil))
}

func main() {
	// Run our health check endpoint in another goroutine
	go healthCheck()

	// Get the gRPC service host and port
	rpcServerHost := os.Getenv("SERVICE_HOST")

	if rpcServerHost == "" {
		rpcServerHost = ":7070"
		log.Printf("Defaulting to host service at %s", rpcServerHost)
	}

	// Bind a TCP listener to the port
	listen, err := net.Listen("tcp", rpcServerHost)
	if err != nil {
		log.Fatalf("Failed to bind TCP listener: %v", err)
	}

	fmt.Println("gRPC server listening at", rpcServerHost)

	// Create a new instance of our ConflictService server that implements the gRPC interface we defined (and then autogenerated the go package for with protoc)
	server := c.NewServer()

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Attach our implementation of the gRPC interface defined in the generated package (the one generated by protoc) and the interface, to the gRPC server
	protobufs.RegisterConflictServiceServer(grpcServer, server)

	// Have the gRPC server listen and serve on the tcp listener created above
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to get gRPC server to serve: $s", err)
	}
}
