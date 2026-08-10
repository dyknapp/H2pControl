package internal

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "h2pcontrol.manager/pb"
)

type ServerEntry struct {
	ID             string
	AdvertisedAddr string
	LastSeen       time.Time
	Metadata       *pb.ServerDefinition
}

type ServerRegistry struct {
	mu      sync.RWMutex
	servers map[string]*ServerEntry
}

func NewServerRegistry() *ServerRegistry {
	return &ServerRegistry{
		servers: make(map[string]*ServerEntry),
	}
}

func newRegistrationID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("Generate registration ID: %w", err)
	}

	return id.String(), nil
}

func (r *ServerRegistry) RegisterServer(
	ctx context.Context,
	in *pb.RegisterRequest) (*pb.RegisterResponse, error) {

	// Use the actual service port
	advertisedAddr := net.JoinHostPort(
		in.Server.GetAdvertisedHost(),
		in.Server.GetPort(),
	)

	// Generate new registration id
	registrationID, err := newRegistrationID()
	if err != nil {
		return nil, err
	}

	entry := &ServerEntry{
		ID:             registrationID,
		AdvertisedAddr: advertisedAddr,
		LastSeen:       time.Now(),
		Metadata:       in.Server,
	}

	log.Printf("Server wants to connect")

	r.mu.Lock()
	r.servers[registrationID] = entry
	r.mu.Unlock()

	SaveProtoFiles(in)
	log.Printf(
		"Registered %s at %s [first 8 chars of id = %s]",
		in.Server.GetServerName(),
		advertisedAddr,
		registrationID[:8],
	)

	return &pb.RegisterResponse{
		Result:         "Server registered successfully",
		RegistrationId: registrationID,
	}, nil
}

func (r *ServerRegistry) FetchServers(ctx context.Context, req *pb.Empty) (*pb.FetchServersResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var serverList []*pb.FetchServerDefinition
	for _, entry := range r.servers {
		// The port in addr is the port of the h2pcontrol server process, not of the running server
		serverList = append(
			serverList,
			&pb.FetchServerDefinition{
				Name:        entry.Metadata.GetServerName(),
				Description: entry.Metadata.GetServerName(),
				Addr:        entry.AdvertisedAddr,
			},
		)
	}

	return &pb.FetchServersResponse{
		Servers: serverList,
	}, nil
}

func (r *ServerRegistry) FetchSpecificServer(ctx context.Context, req *pb.FetchSpecificServerRequest) (*pb.FetchSpecificServerResponse, error) {
	r.mu.RLock()
	svc, ok := r.servers[req.GetAddr()]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("server with addr %s not found", req.GetAddr())
	}

	if entry, ok := r.servers[req.GetAddr()]; ok {
		tmpBase := os.TempDir()
		proto_path := filepath.Join(
			tmpBase,
			"h2pcontrol_proto",
			entry.Metadata.GetServerName(),
			entry.Metadata.GetVersion(),
		)

		proto_files, err := os.ReadDir(proto_path)
		if err != nil {
			log.Fatal("Unable to read proto dir")
		}

		var proto_string string
		for _, proto_file := range proto_files {
			content, err := os.ReadFile(filepath.Join(proto_path, proto_file.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read proto file %s: %v", proto_file.Name(), err)
			}
			proto_string += string(content)
		}

		if err != nil {
			return nil, fmt.Errorf("could not find server function definition for address %s", req.GetAddr())
		}

		return &pb.FetchSpecificServerResponse{
			ServerDefinition: &pb.FetchServerDefinition{
				Name:        svc.Metadata.GetServerName(),
				Description: svc.Metadata.GetServerName(),
				Addr:        req.GetAddr(),
			},
			Proto: proto_string,
		}, nil
	}
	return nil, fmt.Errorf("something went wrong fetching server %s", req.GetAddr())
}

func (r *ServerRegistry) RemoveServer(registrationID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.servers[registrationID]
	if !exists {
		return false
	}

	delete(r.servers, registrationID)
	return true
}

func (r *ServerRegistry) UpdateHeartbeat(registrationID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.servers[registrationID]

	if !exists {
		return false
	}
	entry.LastSeen = time.Now()
	return true
}

func SaveProtoFiles(in *pb.RegisterRequest) error {

	tmpBase := os.TempDir()
	dirPath := filepath.Join(
		tmpBase,
		"h2pcontrol_proto",
		in.Server.GetServerName(),
		in.Server.GetVersion(),
	)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	for _, file := range in.Server.ProtoFiles {
		if err := os.WriteFile(filepath.Join(dirPath, file.Name), file.Content, 0644); err != nil {
			return err
		}
	}
	return nil
}

func (r *ServerRegistry) RemoveExpired(
	cutoff time.Time,
) []*ServerEntry {
	r.mu.Lock()
	defer r.mu.Unlock()

	var expired []*ServerEntry

	for registrationID, entry := range r.servers {
		if entry.LastSeen.Before(cutoff) {
			expired = append(expired, entry)
			delete(r.servers, registrationID)
		}
	}

	return expired
}
