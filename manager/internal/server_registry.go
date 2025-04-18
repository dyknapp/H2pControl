package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"h2pcontrol.manager/helper"
	pb "h2pcontrol.manager/pb"
)

type ServerEntry struct {
	LastSeen  time.Time
	Metadata  *pb.ServerDefinition
	Heartbeat chan struct{}
}

type ServerRegistry struct {
	mu      sync.RWMutex
	servers map[string]*ServerEntry // or your own struct
}

func NewServerRegistry() *ServerRegistry {
	return &ServerRegistry{
		servers: make(map[string]*ServerEntry),
	}
}

func (r *ServerRegistry) RegisterServer(ctx context.Context, in *pb.RegisterRequest, addr string) (*pb.RegisterResponse, error) {
	entry := &ServerEntry{
		LastSeen:  time.Now(),
		Metadata:  in.Server,
		Heartbeat: make(chan struct{}),
	}

	r.mu.Lock()
	r.servers[addr] = entry
	r.mu.Unlock()

	log.Printf("Server connected: '%v' running '%v.%v'", addr, in.Server.GetServerName(), in.Server.GetVersion())
	SaveProtoFiles(in)

	return &pb.RegisterResponse{
		Result: "Server registered successfully",
	}, nil
}

func (r *ServerRegistry) FetchServers(ctx context.Context, req *pb.Empty) (*pb.FetchServersResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var serverList []*pb.FetchServerDefinition
	for addr, svc := range r.servers {
		// The port in addr is the port of the h2pcontrol server process, not of the running server
		ip, _ := helper.SplitAddr(addr)
		server := pb.FetchServerDefinition{
			Name:        svc.Metadata.GetServerName(),
			Description: svc.Metadata.GetServerName(),
			Addr:        ip + ":" + svc.Metadata.Port,
		}
		serverList = append(serverList, &server)
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
		proto_path := filepath.Join("proto", entry.Metadata.GetServerName(), entry.Metadata.GetVersion())

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

func (r *ServerRegistry) UpdateHeartbeat(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok := r.servers[addr]; ok {
		entry.LastSeen = time.Now()
	}
}

func (r *ServerRegistry) MonitorHeartbeats() {

	for {
		r.mu.RLock()
		for addr, entry := range r.servers {
			if time.Since(entry.LastSeen) > 30*time.Second {
				delete(r.servers, addr)
				log.Printf("Server '%v' running '%v.%v' DISCONNECTED: did not send heartbeat in past 30 seconds", addr, entry.Metadata.ServerName, entry.Metadata.Version)
			} else if time.Since(entry.LastSeen) > 2*time.Second {
				log.Printf("Server '%v' running '%v.%v' did not respond to heartbeat", addr, entry.Metadata.ServerName, entry.Metadata.Version)
			} else {
				log.Printf("Server '%v' running '%v.%v' still alive", addr, entry.Metadata.ServerName, entry.Metadata.Version)

			}
		}
		r.mu.RUnlock()
		time.Sleep(5 * time.Second)
	}

}

func SaveProtoFiles(in *pb.RegisterRequest) error {

	dirPath := filepath.Join(
		"proto",
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
