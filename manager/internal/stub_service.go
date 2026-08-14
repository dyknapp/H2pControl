package internal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "h2pcontrol.manager/pb"
)

type StubService struct {
	registry *ServerRegistry
}

type StubSource struct {
	ServerName  string
	Version     string
	ProtoSHA256 string
	ProtoFiles  []*pb.File
}

type WheelArtifact struct {
	Filename string
	Data     []byte
	SHA256   string
}

var pythonDistributionSeparator = regexp.MustCompile(`[-_.]+`)

func createPyProjectToml(packageDir string, name string, version string) error {
	config := map[string]interface{}{
		"build-system": map[string]interface{}{
			"requires":      []string{"setuptools", "wheel"},
			"build-backend": "setuptools.build_meta",
		},
		"project": map[string]interface{}{
			"name":            name,
			"version":         version,
			"description":     "Generated proto package",
			"requires-python": ">=3.11",
			"dependencies": []string{
				"betterproto2>=0.4.0",
			},
		},
		"tool": map[string]interface{}{
			"setuptools": map[string]interface{}{
				"package-dir": map[string]interface{}{
					"": "src",
				},
				"packages": map[string]interface{}{
					"find": map[string]interface{}{
						"where": []string{"src"},
					},
				},
			},
		},
	}

	content, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal pyproject.toml: %w", err)
	}
	if err := os.WriteFile(
		filepath.Join(packageDir, "pyproject.toml"),
		content,
		0o644,
	); err != nil {
		return fmt.Errorf("write pyproject.toml: %w", err)
	}

	return nil
}

func buildPythonWheel(
	ctx context.Context,
	source *StubSource,
) (*WheelArtifact, error) {
	if source == nil {
		return nil, errors.New("stub source is required")
	}
	if source.ServerName == "" {
		return nil, errors.New("server name is required")
	}
	if source.Version == "" {
		return nil, errors.New("server version is required")
	}
	if len(source.ProtoFiles) == 0 {
		return nil, errors.New("at least one proto file is required")
	}

	distributionName := strings.Trim(
		pythonDistributionSeparator.ReplaceAllString(
			strings.ToLower(source.ServerName),
			"-",
		),
		"-",
	)
	if distributionName == "" {
		return nil, errors.New("server name cannot form a Python package name")
	}
	distributionName += "-proto"
	packageName := strings.ReplaceAll(distributionName, "-", "_")

	projectDir, err := os.MkdirTemp("", "h2pcontrol-python-wheel-")
	if err != nil {
		return nil, fmt.Errorf("create temporary wheel project: %w", err)
	}
	defer os.RemoveAll(projectDir)

	protoDir := filepath.Join(projectDir, "proto")
	packageDir := filepath.Join(projectDir, "src", packageName)
	distDir := filepath.Join(projectDir, "dist")
	for _, dir := range []string{protoDir, packageDir, distDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create build directory %q: %w", dir, err)
		}
	}

	if err := createPyProjectToml(
		projectDir,
		distributionName,
		source.Version,
	); err != nil {
		return nil, err
	}
	if err := os.WriteFile(
		filepath.Join(packageDir, "__init__.py"),
		nil,
		0o644,
	); err != nil {
		return nil, fmt.Errorf("create package initializer: %w", err)
	}

	protoNames := make([]string, 0, len(source.ProtoFiles))
	for _, protoFile := range source.ProtoFiles {
		if protoFile == nil {
			return nil, errors.New("proto file is required")
		}

		name := filepath.Clean(filepath.FromSlash(protoFile.GetName()))
		if name == "." || filepath.IsAbs(name) ||
			name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("invalid proto file name %q", protoFile.GetName())
		}

		path := filepath.Join(protoDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create directory for proto file %q: %w", name, err)
		}
		if err := os.WriteFile(path, protoFile.GetContent(), 0o644); err != nil {
			return nil, fmt.Errorf("write proto file %q: %w", name, err)
		}
		protoNames = append(protoNames, filepath.ToSlash(name))
	}

	compileArgs := []string{
		"run",
		"--no-project",
		"--with", "grpcio-tools",
		"--with", "betterproto2-compiler==0.4.0",
		"python",
		"-m", "grpc_tools.protoc",
		"--proto_path", protoDir,
		"--python_betterproto2_out=" + packageDir,
	}
	compileArgs = append(compileArgs, protoNames...)
	compileCmd := exec.CommandContext(ctx, "uv", compileArgs...)
	compileCmd.Dir = protoDir
	if output, err := compileCmd.CombinedOutput(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("compile Python stubs: %w", ctxErr)
		}
		return nil, fmt.Errorf(
			"compile Python stubs with uv: %w: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	buildCmd := exec.CommandContext(
		ctx,
		"uv",
		"build",
		"--wheel",
		"--out-dir", distDir,
		projectDir,
	)
	buildCmd.Dir = projectDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("build Python wheel: %w", ctxErr)
		}
		return nil, fmt.Errorf(
			"build Python wheel with uv: %w: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	entries, err := os.ReadDir(distDir)
	if err != nil {
		return nil, fmt.Errorf("read wheel output directory: %w", err)
	}
	var wheelPath string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".whl" {
			continue
		}
		if wheelPath != "" {
			return nil, errors.New("uv build produced more than one wheel")
		}
		wheelPath = filepath.Join(distDir, entry.Name())
	}
	if wheelPath == "" {
		return nil, errors.New("uv build did not produce a wheel")
	}

	wheelData, err := os.ReadFile(wheelPath)
	if err != nil {
		return nil, fmt.Errorf("read built wheel: %w", err)
	}
	wheelSHA256 := sha256.Sum256(wheelData)

	return &WheelArtifact{
		Filename: filepath.Base(wheelPath),
		Data:     wheelData,
		SHA256:   fmt.Sprintf("%x", wheelSHA256),
	}, nil
}

func NewStubService(registry *ServerRegistry) *StubService {
	return &StubService{
		registry: registry,
	}
}

func (r *ServerRegistry) GetStubSource(
	serverName string,
	version string,
) (*StubSource, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, entry := range r.servers {
		if entry.Metadata.GetServerName() != serverName ||
			entry.Metadata.GetVersion() != version {
			continue
		}

		protoFiles := make([]*pb.File, len(entry.Metadata.GetProtoFiles()))
		for i, file := range entry.Metadata.GetProtoFiles() {
			protoFiles[i] = &pb.File{
				Name:    file.GetName(),
				Content: bytes.Clone(file.GetContent()),
			}
		}

		return &StubSource{
			ServerName:  entry.Metadata.GetServerName(),
			Version:     entry.Metadata.GetVersion(),
			ProtoSHA256: entry.ProtoSha256,
			ProtoFiles:  protoFiles,
		}, true
	}

	return nil, false
}

func (s *StubService) GetStub(
	ctx context.Context,
	in *pb.StubRequest,
) (*pb.StubResponse, error) {
	serverName := in.GetServerName()
	version := in.GetVersion()
	language := in.GetLanguage()

	if serverName == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"server name is required",
		)
	}
	if version == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"server version is required",
		)
	}
	if language != "python" {
		return nil, status.Errorf(
			codes.Unimplemented,
			"language %q is not supported",
			language,
		)
	}

	source, found := s.registry.GetStubSource(serverName, version)
	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"server %s version %s is not online",
			serverName,
			version,
		)
	}

	artifact, err := buildPythonWheel(ctx, source)
	if err != nil {
		log.Printf(
			"Build stub for %s version %s: %v",
			serverName,
			version,
			err,
		)
		return nil, status.Error(
			codes.Internal,
			"failed to build stub wheel",
		)
	}

	return &pb.StubResponse{
		Filename:    artifact.Filename,
		WheelData:   artifact.Data,
		WheelSha256: artifact.SHA256,
		ProtoSha256: source.ProtoSHA256,
	}, nil
}

func compileProtoHandler(in *pb.StubRequest, proto_path string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "h2pcontrol-")
	if err != nil {
		log.Fatal("Error creating temp dir:", err)
	}

	proto_files, err := os.ReadDir(proto_path)
	if err != nil {
		log.Fatal("Unable to read proto dir")
	}

	if in.Language == "python" {

		for _, proto_file := range proto_files {
			// Have to run this through bash to make sure it is in the same env, otherwise grpc_tools will not be available
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
