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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
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
	return s.registry.RegisterServer(ctx, in)
}

func (s *server) UnregisterServer(
	ctx context.Context,
	in *pb.UnregisterServerRequest,
) (*pb.Empty, error) {
	registrationID := in.GetRegistrationId()
	if registrationID == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"Registration ID is required!",
		)
	}

	removed := s.registry.RemoveServer(registrationID)
	if removed {
		log.Printf(
			"Removed server with registration ID %s",
			registrationID,
		)
	}

	return &pb.Empty{}, nil
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
	var registrationID string

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
			log.Printf(
				"Heartbeat stream closed from %s...: %v",
				registrationID[:8],
				err,
			)

			if registrationID != "" {
				s.registry.RemoveServer(registrationID)
			}

			close(done)
			return nil
		}

		registrationID = ping.GetRegistrationId()

		if !s.registry.UpdateHeartbeat(registrationID) {
			return status.Errorf(
				codes.NotFound,
				"registration %s not found",
				registrationID,
			)
		}

		if registrationID != "" {
			log.Printf(
				"Received heartbeat from %s... at %v",
				registrationID[:8],
				time.Unix(ping.Timestamp, 0),
			)
		}
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

	const (
		heartbeatTimeout = 6 * time.Second // Assume server is dead if it doesn't respond for this long
		reaperInterval   = 1 * time.Second // At this interval, go through and officially delete non-responsive servers
	)

	go func() {
		ticker := time.NewTicker(reaperInterval)
		defer ticker.Stop()
		for now := range ticker.C {
			cutoff := now.Add(-heartbeatTimeout)

			for _, entry := range srv.registry.RemoveExpired(cutoff) {
				log.Printf(
					"Registration expired: %s at %s",
					entry.Metadata.GetServerName(),
					entry.AdvertisedAddr,
				)
			}
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
