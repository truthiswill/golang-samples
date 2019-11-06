// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START run_grpc_server_multiplex]

// Sample relay acts as an intermediary to the ping service.
package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	pb "github.com/GoogleCloudPlatform/golang-samples/run/grpc-ping/pkg/api/v1"
)

// Conn holds an open connection to the ping service.
var Conn *grpc.ClientConn

func main() {
	log.Printf("grpg-ping: starting server...")

	if os.Getenv("GRPC_PING_HOST") == "" {
		log.Fatal("Must specify a ping host with 'GRPC_PING_HOST' environment variable. E.g., example.com")
	}

	var err error
	Conn, err = NewConn(os.Getenv("GRPC_PING_HOST"), os.Getenv("GRPC_PING_PORT"), os.Getenv("GRPC_PING_INSECURE") != "")
	if err != nil {
		log.Fatal(err)
	}

	// Determine port for gRPC service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("net.Listen: %v", err)
	}

	// Create a connection multiplexer, allowing the same port
	// to be used for both gRPC and HTTP traffic.
	mux := cmux.New(listener)

	// Match connections in order:
	// First grpc, then HTTP, and otherwise Go RPC/TCP.
	httpListener := mux.Match(cmux.HTTP1Fast())
	grpcListener := mux.Match(cmux.Any())

	httpServer := &http.Server{
		Handler: http.HandlerFunc(indexHandler),
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPingServiceServer(grpcServer, &pingService{})

	group := errgroup.Group{}
	group.Go(func() error { return httpServer.Serve(httpListener) })
	group.Go(func() error { return grpcServer.Serve(grpcListener) })
	group.Go(func() error { return mux.Serve() })
	log.Fatal(group.Wait())
}

// [END run_grpc_server_multiplex]
