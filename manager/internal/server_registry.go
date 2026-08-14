package internal

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "h2pcontrol.manager/pb"
)

// TODO:
// - Make file writes atomic
// - Failures with json file can look like errors from hash differences (version wasn't changed)

type ServerEntry struct {
	ID             string
	AdvertisedAddr string
	LastSeen       time.Time
	Metadata       *pb.ServerDefinition
	ProtoSha256    string
}

type serverVersionKey struct {
	ServerName string
	Version    string
}

type ServerRegistry struct {
	mu      sync.RWMutex
	servers map[string]*ServerEntry
	catalog *ProtoCatalog
}

type ProtoCatalog struct {
	mu          sync.RWMutex
	protoHashes map[serverVersionKey]string
	path        string
}

func NewServerRegistry(catalog *ProtoCatalog) *ServerRegistry {
	return &ServerRegistry{
		servers: make(map[string]*ServerEntry),
		catalog: catalog,
	}
}

func NewProtoCatalog(path string) (*ProtoCatalog, error) {
	c := &ProtoCatalog{
		protoHashes: make(map[serverVersionKey]string),
		path:        path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File hasn't been created yet.  Return an empty catalog
			return c, nil
		}
		return nil, fmt.Errorf("Read catalog file: %w", err)
	}

	var fileData protoHashFile
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("Parse catalog JSON: %w", err)
	}

	for _, record := range fileData.Records {
		key := serverVersionKey{
			ServerName: record.ServerName,
			Version:    record.Version,
		}
		c.protoHashes[key] = record.ProtoSHA256
	}

	return c, nil
}

func (c *ProtoCatalog) saveLocked() error {
	fileData := protoHashFile{
		FormatVersion: 1,
		Records:       make([]protoHashRecord, 0, len(c.protoHashes)),
	}

	for key, hash := range c.protoHashes {
		fileData.Records = append(fileData.Records, protoHashRecord{
			ServerName:  key.ServerName,
			Version:     key.Version,
			ProtoSHA256: hash,
		})
	}

	data, err := json.MarshalIndent(fileData, "", " ")
	if err != nil {
		return fmt.Errorf("Json loading: %w", err)
	}

	// Make sure the directory exists before writing
	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return fmt.Errorf("Create protocatalog directory: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0644); err != nil {
		return fmt.Errorf("Write protocatalog file: %w", err)
	}

	return nil
}

func (c *ProtoCatalog) CheckAndRecord(
	key serverVersionKey,
	protoSHA256 string,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	existingHash, exists := c.protoHashes[key]
	if exists {
		if existingHash != protoSHA256 {
			return fmt.Errorf(
				"Server %s version %s has proto has %s, previously %s.  Try updating the server version if you have modified the protobuf.",
				key.ServerName,
				key.Version,
				protoSHA256,
				existingHash,
			)
		}

		// This protobuf already exists, and the new registered one is consistent qwith it
		return nil
	}

	// This protobuf is unknown, but it also doesn't clash with an existing one.  Add it as a new one.
	c.protoHashes[key] = protoSHA256
	if err := c.saveLocked(); err != nil {
		// If we couldn't save the file, roll back the change.
		delete(c.protoHashes, key)
		return fmt.Errorf("Save protocatalog: %w", err)
	}

	return nil
}

func newRegistrationID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("Generate registration ID: %w", err)
	}

	return id.String(), nil
}

type protoHashFile struct {
	FormatVersion int               `json:"format_version"`
	Records       []protoHashRecord `json:"records"`
}

type protoHashRecord struct {
	ServerName  string `json:"server_name"`
	Version     string `json:"version"`
	ProtoSHA256 string `json:"proto_sha256"`
}

var validProtoFilename = regexp.MustCompile(`^[^/\\]+\.proto$`)

func hashProtoFiles(files []*pb.File) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("At least one proto file is required.")
	}

	// Make sure proto files aren't nil
	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		name := file.GetName()

		if !validProtoFilename.MatchString(name) {
			return "", fmt.Errorf("invalid proto filename: %q", name)
		}

		if _, exists := seen[name]; exists {
			return "", fmt.Errorf("duplicate proto filename: %q", name)
		}

		seen[name] = struct{}{}
	}

	// Avoid mutating the input.  Alphabetical order
	ordered := append([]*pb.File(nil), files...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].GetName() < ordered[j].GetName()
	})

	hasher := sha256.New()
	var lengthBuffer [8]byte

	writePart := func(data []byte) {
		binary.BigEndian.PutUint64(lengthBuffer[:], uint64(len(data)))
		hasher.Write(lengthBuffer[:])
		hasher.Write(data)
	}

	for _, file := range ordered {
		writePart([]byte(file.GetName()))
		writePart(file.GetContent())
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func (r *ServerRegistry) RegisterServer(
	ctx context.Context,
	in *pb.RegisterRequest) (*pb.RegisterResponse, error) {

	server := in.GetServer()

	// Validation of server registration request
	if server == nil {
		return nil, status.Error(codes.InvalidArgument, "Server definition cannot be nil.")
	}
	if server.GetServerName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Server name cannot be empty.")
	}
	if server.GetVersion() == "" {
		return nil, status.Error(codes.InvalidArgument, "Server version cannot be empty.")
	}

	// Hash the server
	protoSHA256, err := hashProtoFiles(server.GetProtoFiles())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid .proto files: %v", err)
	}
	key := serverVersionKey{
		ServerName: server.GetServerName(),
		Version:    server.GetVersion(),
	}

	if err := r.catalog.CheckAndRecord(key, protoSHA256); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// Use the actual service port
	advertisedAddr := net.JoinHostPort(
		server.GetAdvertisedHost(),
		server.GetPort(),
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
		Metadata:       server,
		ProtoSha256:    protoSHA256,
	}

	log.Printf("Server wants to connect")

	r.mu.Lock()
	r.servers[registrationID] = entry
	r.mu.Unlock()

	log.Printf(
		"Registered %s at %s [first 8 chars of id = %s]",
		in.Server.GetServerName(),
		advertisedAddr,
		registrationID[:8],
	)

	return &pb.RegisterResponse{
		Result:         "Server registered successfully",
		RegistrationId: registrationID,
		ProtoSha256:    protoSHA256,
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
				Version:     entry.Metadata.GetVersion(),
				ProtoSha256: entry.ProtoSha256,
			},
		)
	}

	return &pb.FetchServersResponse{
		Servers: serverList,
	}, nil
}

func (r *ServerRegistry) FetchSpecificServer(ctx context.Context, req *pb.FetchSpecificServerRequest) (*pb.FetchSpecificServerResponse, error) {
	r.mu.RLock()
	var server_entry *ServerEntry
	for _, entry := range r.servers {
		if entry.AdvertisedAddr == req.GetAddr() {
			server_entry = entry
			break
		}
	}
	r.mu.RUnlock()

	if server_entry == nil {
		return nil, fmt.Errorf("Server with address %s not found.", req.GetAddr())
	}

	tmpBase := os.TempDir()
	proto_path := filepath.Join(
		tmpBase,
		"h2pcontrol_proto",
		server_entry.Metadata.GetServerName(),
		server_entry.Metadata.GetVersion(),
	)

	proto_files, err := os.ReadDir(proto_path)
	if err != nil {
		return nil, fmt.Errorf("unable to read proto dir: %w", err)
	}

	var proto_string string
	for _, proto_file := range proto_files {
		content, err := os.ReadFile(filepath.Join(proto_path, proto_file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read proto file %s: %v", proto_file.Name(), err)
		}
		proto_string += string(content)
	}

	return &pb.FetchSpecificServerResponse{
		ServerDefinition: &pb.FetchServerDefinition{
			Name:        server_entry.Metadata.GetServerName(),
			Description: server_entry.Metadata.GetServerName(),
			Addr:        req.GetAddr(),
		},
		Proto: proto_string,
	}, nil
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

func DefaultProtoCatalogPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appConfigDir := filepath.Join(configDir, "h2pcontrol-manager", "proto-catalog.json")
	log.Printf("Reminder: proto catalog json file is saved at: %s", appConfigDir)
	return appConfigDir, nil
}

func SaveProtoHash() (string, error) {
	appConfigDir, err := DefaultProtoCatalogPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}

	configFile := filepath.Join(appConfigDir, "proto-catalog.json")

	return configFile, nil
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
