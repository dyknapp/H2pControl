package cmd

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	pb "h2pcontrol.client/pb"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func Run(c pb.ManagerClient, ctx context.Context, runCommand string, service pb.ServiceDefinition, proto_path string) {
	RegisterService(c, ctx, service, proto_path)

	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd, err := startCommand(cmdCtx, runCommand)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer cmd.Wait()

	go runHeartbeat(c)

	waitForShutdown()
	cancel()
}

func startCommand(ctx context.Context, command string) (*exec.Cmd, error) {
	args := strings.Fields(command)
	if len(args) < 2 {
		return nil, fmt.Errorf("invalid command format: need 'shell command'")
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	go streamOutput(stdout, "OUT:")
	go streamOutput(stderr, "ERR:")

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start command: %w", err)
	}

	return cmd, nil
}

func streamOutput(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Printf("%s %s\n", prefix, scanner.Text())
	}
}

func waitForShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Server is running. Press Ctrl+C to stop.")

	sig := <-signalChan
	log.Printf("Received signal: %s. Shutting down...\n", sig)
}

func RegisterService(c pb.ManagerClient, ctx context.Context, service pb.ServiceDefinition, proto_dir_path string) {

	dirEntries, err := os.ReadDir(proto_dir_path)
	if err != nil {
		log.Fatalf("Unable to read proto dir: %v", err)
	}

	for _, entry := range dirEntries {
		file_content, err := os.ReadFile(filepath.Join(proto_dir_path, entry.Name()))
		if err != nil {
			log.Fatalf("Unable to read proto file: %v", err)
		}
		service.ProtoFiles = append(service.ProtoFiles, &pb.File{Name: entry.Name(), Content: file_content})
	}

	request := pb.RegisterRequest{Service: &service}

	r, err := c.RegisterServer(ctx, &request)
	if err != nil {
		// Make this error handling nicer
		log.Fatalf("Unable to connect to h2pcontrol Manager, is it running? %v", err)
	}
	log.Println(r.Result)

}

func GetActiveServices(c pb.ManagerClient, ctx context.Context) {
	request := pb.ActivateServicesRequest{}
	r, err := c.GetActiveServices(ctx, &request)
	if err != nil {
		log.Fatalf("Unable to retreive other services from manager")
	}
	log.Println(r)
}

func runHeartbeat(client pb.ManagerClient) {
	for {

		_, err := client.Heartbeat(context.Background(), &pb.HeartbeatPing{}, grpc.EmptyCallOption{})
		if err != nil {
			log.Fatalf("Failed to start heartbeat stream: %v", err)
		}

		if err != nil {
			log.Fatalf("Failed to receive heartbeat response: %v", err)
		}

		// log.Printf("Received pong from server: %v", pong.Healthy)

		time.Sleep(1 * time.Second)
	}
}

func GetStubs(c pb.ManagerClient, ctx context.Context, dependencies []pb.ServiceDefinition, language string) {

	for _, dependency := range dependencies {
		GetStub(c, ctx, dependency.ServiceName, dependency.Version, language)
	}
}

func extractZipData(zipData []byte, outputDir string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return err
	}

	for _, file := range zipReader.File {
		targetPath := filepath.Join(outputDir, filepath.Clean(file.Name))
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		if err := extractFile(file, targetPath); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(zipFile *zip.File, targetPath string) error {
	srcFile, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}

func GetStub(c pb.ManagerClient, ctx context.Context, service_name string, version string, language string) {
	r, err := c.GetStub(ctx, &pb.StubRequest{ServiceName: service_name, Version: version, Language: language})
	if err != nil {
		log.Fatalf("could not get stub file: %v", err)
	}

	dirPath := filepath.Join("stubs",
		service_name,
		language,
		version)

	if err := extractZipData(r.ZipData, dirPath); err != nil {
		log.Fatalf("Not a valid zip file %s", err)
	}

	log.Printf("Finished receiving stubs")
}
