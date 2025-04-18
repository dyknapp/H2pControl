package grpc

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"h2pcontrol.manager/internal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
	pb "h2pcontrol.manager/pb"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedManagerServer
	sync.RWMutex
	registry     *internal.ServerRegistry
	stub_service *internal.StubService
}

// func (s *server) RegisterServer(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
// 	peer, _ := peer.FromContext(ctx)
// 	addr := peer.Addr.String()

// 	entry := &serviceEntry{
// 		lastSeen:  time.Now(),
// 		metadata:  in.Server,
// 		heartbeat: make(chan struct{}),
// 	}

// 	s.Lock()
// 	s.activeServers[addr] = entry
// 	s.Unlock()

// 	log.Printf("Server connected: '%v' running '%v.%v'", addr, in.Server.GetServerName(), in.Server.GetVersion())

// 	var dirPath = filepath.Join(
// 		"proto",
// 		in.Server.ServerName,
// 		in.Server.Version)

// 	os.MkdirAll(dirPath, 0755)

// 	for _, file := range in.Server.ProtoFiles {
// 		os.WriteFile(filepath.Join(dirPath, file.Name), file.Content, 0644)
// 	}

// 	return &pb.RegisterResponse{
// 		Result: "Server registered successfully",
// 	}, nil
// }

func (s *server) GetStub(ctx context.Context, in *pb.StubRequest) (*pb.StubResponse, error) {
	return s.stub_service.GetStub(ctx, in)
}

func (s *server) RegisterServer(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	peerInfo, _ := peer.FromContext(ctx)
	addr := peerInfo.Addr.String()
	return s.registry.RegisterServer(ctx, in, addr)
}

func (s *server) FetchServers(ctx context.Context, in *pb.Empty) (*pb.FetchServersResponse, error) {
	return s.registry.FetchServers(ctx, in)
}

func (s *server) FetchSpecificServer(ctx context.Context, in *pb.FetchSpecificServerRequest) (*pb.FetchSpecificServerResponse, error) {
	return s.registry.FetchSpecificServer(ctx, in)
}

func (s *server) Heartbeat(ctx context.Context, in *pb.Empty) (*pb.HeartbeatPong, error) {
	peer, _ := peer.FromContext(ctx)
	addr := peer.Addr.String()
	s.registry.UpdateHeartbeat(addr)
	log.Printf("Heartbeat from %v", addr)
	return &pb.HeartbeatPong{Healthy: true}, nil
}

func RunServer() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := &server{
		registry:     internal.NewServerRegistry(),
		stub_service: internal.NewStubService(),
	}

	serverOpts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	s := grpc.NewServer(serverOpts...)
	pb.RegisterManagerServer(s, srv)
	log.Printf("server listening at %v", lis.Addr())

	go srv.registry.MonitorHeartbeats()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
