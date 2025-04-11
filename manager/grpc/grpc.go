package grpc

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	pb "otter.manager/pb"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type serviceEntry struct {
	lastSeen  time.Time
	metadata  *pb.ServiceDefinition
	heartbeat chan struct{}
}

type server struct {
	pb.UnimplementedManagerServer
	activeServices map[string]*serviceEntry
	sync.RWMutex
}

func (s *server) GetStub(_ context.Context, in *pb.StubRequest) (*pb.StubResponse, error) {
	log.Printf("Received call for service: '%v %v' for '%v'", in.GetServiceName(), in.GetVersion(), in.GetLanguage())

	dirPath := filepath.Join(
		"services",
		in.GetServiceName(),
		in.GetVersion(),
		in.GetLanguage(),
	)
	log.Printf("Serving file from %v", dirPath)

	dirContents, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Service %s/%s/%s not found",
			in.GetServiceName(),
			in.GetVersion(),
			in.GetLanguage())
	}

	var responseFiles []*pb.File
	for _, entry := range dirContents {

		filePath := filepath.Join(dirPath, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to read file: %v", err)
		}

		responseFiles = append(responseFiles, &pb.File{
			Name:    entry.Name(),
			Content: content,
		})
	}

	if len(responseFiles) == 0 {
		return nil, status.Errorf(codes.NotFound, "No valid stub files found")
	}
	return &pb.StubResponse{
		Files: responseFiles,
	}, nil
}

func (s *server) RegisterServer(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	peer, _ := peer.FromContext(ctx)
	addr := peer.Addr.String()

	entry := &serviceEntry{
		lastSeen:  time.Now(),
		metadata:  in.Service,
		heartbeat: make(chan struct{}),
	}

	s.Lock()
	s.activeServices[addr] = entry
	s.Unlock()

	log.Printf("Service connected: '%v' running '%v.%v'", addr, in.Service.GetServiceName(), in.Service.GetVersion())

	return &pb.RegisterResponse{
		Result: "Server registered successfully",
	}, nil
}

func (s *server) Heartbeat(ctx context.Context, in *pb.HeartbeatPing) (*pb.HeartbeatPong, error) {
	peer, _ := peer.FromContext(ctx)
	addr := peer.Addr.String()

	s.Lock()
	entry := s.activeServices[addr]
	entry.lastSeen = time.Now()
	s.Unlock()

	log.Printf("Heartbeat from %v", addr)

	return &pb.HeartbeatPong{
		Healthy: true,
	}, nil
}

func (s *server) GetActiveServices(ctx context.Context, in *pb.ActivateServicesRequest) (*pb.ActiveServicesResponse, error) {
	s.RLock()
	defer s.RUnlock()

	results := make([]*pb.ServiceDefinition, 0, len(s.activeServices))
	for _, entry := range s.activeServices {
		if entry.metadata != nil {
			results = append(results, entry.metadata)
		}
	}
	return &pb.ActiveServicesResponse{Services: results}, nil
}

func RunServer() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := &server{
		activeServices: make(map[string]*serviceEntry),
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

	go srv.monitorHeartbeats()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) monitorHeartbeats() {

	for {
		s.RLock()
		for addr, entry := range s.activeServices {
			if time.Since(entry.lastSeen) > 30*time.Second {
				delete(s.activeServices, addr)
				log.Printf("Service '%v' running '%v.%v' DISCONNECTED: did not send heartbeat in past 30 seconds", addr, entry.metadata.ServiceName, entry.metadata.Version)
			} else if time.Since(entry.lastSeen) > 2*time.Second {
				log.Printf("Service '%v' running '%v.%v' did not respond to heartbeat", addr, entry.metadata.ServiceName, entry.metadata.Version)
			} else {
				log.Printf("Service '%v' running '%v.%v' still alive", addr, entry.metadata.ServiceName, entry.metadata.Version)

			}

		}
		s.RUnlock()
		time.Sleep(5 * time.Second)
	}

}
