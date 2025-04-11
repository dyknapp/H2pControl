package grpc

import (
	"archive/zip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
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

	proto_path := filepath.Join("proto", in.GetServiceName(), in.GetVersion())
	dirPath, err := compileProtoHandler(in, proto_path)

	if err != nil {
		println("Could not compile proto handler")
		return nil, err
	}

	buf, err := createZip(dirPath)
	if err != nil {
		println("Could not create zip")
		return nil, err
	}

	os.WriteFile("test_zip.zip", buf, 0644)

	return &pb.StubResponse{
		ZipData: buf,
		Name:    filepath.Base(dirPath),
	}, nil
}

// Bit of a long ugly function..
func createZip(sourceDir string) ([]byte, error) {
	srcInfo, err := os.Stat(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("source directory error: %w", err)
	}
	if !srcInfo.IsDir() {
		return nil, errors.New("source must be a directory")
	}

	zipFile, err := os.CreateTemp("", "tmpfile-")

	if err != nil {
		return nil, fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()
	defer os.Remove(zipFile.Name())

	zipWriter := zip.NewWriter(zipFile)

	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("relative path error: %w", err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("header creation error: %w", err)
		}
		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		entryWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("entry creation error: %w", err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("file open error: %w", err)
		}
		defer file.Close()

		_, err = io.Copy(entryWriter, file)
		if err != nil {
			return fmt.Errorf("file copy error: %w", err)
		}

		return nil
	})

	zipWriter.Close()

	// inefficient to write and then read but fine for now.
	zipContent, err := os.ReadFile(zipFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %w", err)
	}
	return zipContent, nil
}

func compileProtoHandler(in *pb.StubRequest, proto_path string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "otter-")
	if err != nil {
		log.Fatal("Error creating temp dir:", err)
	}

	proto_files, err := os.ReadDir(proto_path)
	if err != nil {
		log.Fatal("Unable to read proto dir")
	}

	if in.Language == "python" {

		for _, proto_file := range proto_files {
			// args := fmt.Sprintf(
			// 	"-m grpc_tools.protoc --python_betterproto2_out=%s -I. %s",
			// 	tmpDir,
			// 	filepath.Join(proto_path, proto_file.Name()),
			// )

			// cmd := exec.Command("python3", args)
			// cmd.Env = os.Environ()
			// // log.Println("ENV PATH:", os.Getenv("PATH"))
			fullCommand := fmt.Sprintf(
				"source ~/.bashrc && python3 -m grpc_tools.protoc --python_betterproto2_out=%s -I%s %s",
				tmpDir,
				proto_path,
				filepath.Join(proto_path, proto_file.Name()),
			)
			cmd := exec.Command("bash", "-c", fullCommand)

			log.Println(cmd.Args)
			output, err := cmd.CombinedOutput()
			if err != nil {
				// error
				log.Printf("STDOUT: %s", string(output))

				log.Printf("Unable to compile: %v", err)
			}
		}

		return tmpDir, nil
	} else {
		return "", fmt.Errorf("Currently only python is supported")
	}

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

	var dirPath = filepath.Join(
		"proto",
		in.Service.ServiceName,
		in.Service.Version)

	os.MkdirAll(dirPath, 0755)

	for _, file := range in.Service.ProtoFiles {
		os.WriteFile(filepath.Join(dirPath, file.Name), file.Content, 0644)
	}

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
