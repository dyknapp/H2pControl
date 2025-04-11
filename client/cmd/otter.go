package cmd

import (
	"bufio"
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
	pb "otter.client/pb"
)

var (
	addr         = flag.String("addr", "localhost:50051", "the address to connect to")
	service_name = flag.String("service_name", "arduino", "Service Name")
	version      = flag.String("version", "v1.0.0", "Package version")
	language     = flag.String("language", "python", "Programming language")
)

func Run(c pb.ManagerClient, ctx context.Context, runCommand string) {
	RegisterService(c, ctx)

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

func RegisterService(c pb.ManagerClient, ctx context.Context) {

	request := pb.RegisterRequest{Service: &pb.ServiceDefinition{ServiceName: *service_name,
		Version: *version,
		File:    &pb.File{Name: "example_file", Content: []byte{0x01, 0x02, 0x03}}, // TEMP
	}}

	r, err := c.RegisterServer(ctx, &request)
	if err != nil {
		// Make this error handling nicer
		log.Fatalf("Unable to connect to Otter Manager, is the  running? %v", err)
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

type Service struct {
	Name     string `mapstructure:"name"`
	Version  string `mapstructure:"version"`
	Language string `mapstructure:"language"`
}

func GetStubs(c pb.ManagerClient, ctx context.Context, dependencies []Service, language string) {

	for _, dependency := range dependencies {
		GetStub(c, ctx, dependency.Name, dependency.Version, language)
	}

}
func GetStub(c pb.ManagerClient, ctx context.Context, service_name string, version string, language string) {
	r, err := c.GetStub(ctx, &pb.StubRequest{ServiceName: service_name, Version: version, Language: language})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	files := r.GetFiles()

	dirPath := filepath.Join("stubs",
		service_name,
		language,
		version)
	for _, file := range files {
		var filePath = filepath.Join(
			dirPath,
			file.Name,
		)

		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			log.Fatalf("Failed to create directories: %v", err)
		}

		err = os.WriteFile(filePath, file.Content, 0644)
		if err != nil {
			log.Fatalf("Tried to write %v, but received error %v", filePath, err)
		}
	}
	log.Printf("Finished receiving stubs")
}
