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

// func (s *server) Heartbeat(ctx context.Context, in *pb.Empty) (*pb.HeartbeatPong, error) {
// 	peer, _ := peer.FromContext(ctx)
// 	addr := peer.Addr.String()
// 	s.registry.UpdateHeartbeat(addr)
// 	log.Printf("Heartbeat from %v", addr)
// 	return &pb.HeartbeatPong{Healthy: true}, nil
// }

func (s *server) Heartbeat(stream pb.Manager_HeartbeatServer) error {
	peerInfo, _ := peer.FromContext(stream.Context())
	addr := peerInfo.Addr.String()

	// Start a goroutine to send heartbeats to the client
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				pong := &pb.HeartbeatPong{
					Healthy:   true,
					Timestamp: time.Now().Unix(),
				}
				if err := stream.Send(pong); err != nil {
					log.Printf("Error sending heartbeat pong: %v", err)
					close(done)
					return
				}
			case <-done:
				return
			}
		}
	}()

	for {
		ping, err := stream.Recv()
		if err != nil {
			log.Printf("Heartbeat stream closed from %v: %v", addr, err)
			s.registry.RemoveServer(addr)
			close(done)
			return nil
		}

		s.registry.UpdateHeartbeat(addr)
		log.Printf("Received heartbeat from %v at %v", addr, time.Unix(ping.Timestamp, 0))
	}
}

func RunServer() {
	flag.Parse()

	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", *port))
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

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
