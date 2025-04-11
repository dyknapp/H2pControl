package cmd

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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

// func init() {
// 	// flag.Parse()
// 	// Set up a connection to the server.
// 	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(keepalive.ClientParameters{
// 		Time:                20 * time.Second,
// 		Timeout:             10 * time.Second,
// 		PermitWithoutStream: true,
// 	}))
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()
// 	c := pb.NewManagerClient(conn)

// 	ctx := context.Background()
// 	RegisterService(c, ctx)

// 	go runHeartbeat(c)

// 	GetActiveServices(c, ctx)
// 	// Keep main alive to maintain connection
// 	select {}

// }

func Run(c pb.ManagerClient, ctx context.Context) {
	RegisterService(c, ctx)

	go runHeartbeat(c)

	waitForShutdown()
}

func waitForShutdown() {
	// Create a channel to listen for OS signals (e.g., SIGINT, SIGTERM)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Server is running. Press Ctrl+C to stop.")

	// Block until a signal is received
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
		log.Fatalf("Unable to connect to server, is the server running? %v", err)
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

		pong, err := client.Heartbeat(context.Background(), &pb.HeartbeatPing{}, grpc.EmptyCallOption{})
		if err != nil {
			log.Fatalf("Failed to start heartbeat stream: %v", err)
		}

		if err != nil {
			log.Fatalf("Failed to receive heartbeat response: %v", err)
		}

		log.Printf("Received pong from server: %v", pong.Healthy)

		time.Sleep(1 * time.Second)
	}
}

func GetStubs(c pb.ManagerClient, ctx context.Context) {
	r, err := c.GetStub(ctx, &pb.StubRequest{ServiceName: *service_name, Version: *version, Language: *language})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	files := r.GetFiles()

	dirPath := filepath.Join("stubs",
		*service_name,
		*language,
		*version)
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
