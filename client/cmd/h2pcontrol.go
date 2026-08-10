package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/phayes/freeport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	pb "h2pcontrol.client/pb"
)

func Run(
	c pb.ManagerClient,
	ctx context.Context,
	runCommand string,
	server *pb.ServerDefinition,
	proto_path string,
	extraArgs []string,
) {

	// Check if the last entry is a file / check its path, then if it is a file and we have 1 arg it should just run that because
	// we can assume it's an executable? We can also check if it's an executable?
	args := strings.Fields(runCommand)

	// Determine the path to watch for the file watcher
	watchedPath := args[len(args)-1]

	if len(args) < 2 {
		fmt.Printf("invalid command format %s: need 'shell command'", args)
		return
	}

	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var cmd *exec.Cmd
	restart := func() {
		if cmd != nil && cmd.Process != nil {
			if runtime.GOOS == "windows" {
				cmd.Process.Signal(os.Interrupt) // Sends CTRL_BREAK_EVENT which we listen to within python
			} else {
				cmd.Process.Signal(syscall.SIGTERM)
			}
			cmd.Wait()
		}
		if server.Port == "" {
			port, err := freeport.GetFreePort()
			if err != nil {
				panic(err)
			}
			args = append(args, "--port")
			args = append(args, strconv.Itoa(port))
			server.Port = strconv.Itoa(port)
		}
		args = append(args, extraArgs...)
		c, err := startCommand(cmdCtx, args)
		if err != nil {
			fmt.Println("Failed to start command:", err)
			return
		}
		cmd = c
	}

	restart()

	// TODO: Make this more general, but for now will do for python
	fileWatcher, err := NewFileWatcher(300*time.Millisecond, restart)
	if err != nil {
		panic(err)
	}
	defer fileWatcher.Close()

	if err := fileWatcher.Watch(watchedPath); err != nil {
		panic(err)
	}

	// Register the server after it has been started, as otherwise it might not register even though it is not available
	// Do note, we are not checking if the python server itself has any errors.
	registrationID, err := RegisterService(c, ctx, server, proto_path)
	if err != nil {
		return
	}

	go runHeartbeat(ctx, c, registrationID, server)

	waitForShutdown(cmd)

	_, err = c.UnregisterServer(
		context.Background(),
		&pb.UnregisterServerRequest{
			RegistrationId: registrationID,
		},
	)
}

func startCommand(ctx context.Context, args []string) (*exec.Cmd, error) {

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

func waitForShutdown(cmd *exec.Cmd) {
	processExited := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Server is running. Press Ctrl+C to stop.")

	go func() {
		processExited <- cmd.Wait()
	}()

	select {
	case sig := <-signalChan:
		log.Printf("Received %s", sig)
		cmd.Process.Signal(os.Interrupt)
	case err := <-processExited:
		log.Printf("Server process exited: %v", err)
	}
}

func RegisterService(
	c pb.ManagerClient,
	ctx context.Context,
	server *pb.ServerDefinition,
	proto_dir_path string) (string, error) {
	if server.GetAdvertisedHost() == "" {
		log.Fatal("Advertised host was not cnofigured")
	}
	log.Printf(
		"Advertising %s:%s",
		server.GetAdvertisedHost(),
		server.GetPort(),
	)

	dirEntries, err := os.ReadDir(proto_dir_path)
	if err != nil {
		log.Fatalf("Unable to read proto dir: %v", err)
	}

	for _, entry := range dirEntries {
		file_content, err := os.ReadFile(filepath.Join(proto_dir_path, entry.Name()))
		if err != nil {
			log.Fatalf("Unable to read proto file: %v", err)
		}
		server.ProtoFiles = append(server.ProtoFiles, &pb.File{Name: entry.Name(), Content: file_content})
	}

	request := pb.RegisterRequest{Server: server}

	log.Println("Going to register to server")
	rpcStart := time.Now()
	response, err := c.RegisterServer(ctx, &request)
	if err != nil {
		log.Fatalf("Unable to connect to h2pcontrol Manager, is it running? %v", err)
	}
	log.Printf("RegisterServer call took %v\n", time.Since(rpcStart))
	log.Println(response.Result)

	return response.GetRegistrationId(), nil
}

func checkServerHealth(
	parent context.Context,
	client healthpb.HealthClient,
) error {
	ctx, cancel := context.WithTimeout(parent, time.Second)
	defer cancel()

	response, err := client.Check(
		ctx,
		&healthpb.HealthCheckRequest{
			Service: "",
		},
	)

	if err != nil {
		return err
	}

	if response.Status != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("Abnormal server status: %s", response.Status)
	}

	return nil
}

// Stop sending pings after this many abnormal counts.
// No mercy: for now let the manager's setting decide on how long to allow the server to hang, with abnormal_counts_limit=1
const abnormal_counts_limit = 1

func runHeartbeat(
	ctx context.Context,
	client pb.ManagerClient,
	registrationID string,
	server *pb.ServerDefinition,
) {
	// This is a bidirectional streaming heartbeat.
	// Using the passed-in context is usually safer than context.Background() here
	stream, err := client.Heartbeat(ctx)
	if err != nil {
		log.Fatalf("Failed to start heartbeat: %v", err)
	}

	serverAddress := net.JoinHostPort(
		server.GetAdvertisedHost(),
		server.GetPort(),
	)

	serverConn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		// Log the error and exit the goroutine instead of returning err
		log.Printf("Failed to connect to local server for health check: %v", err)
		return
	}
	defer serverConn.Close()

	healthClient := healthpb.NewHealthClient(serverConn)

	// Send heartbeats
	abnormal_counts := 0
	go func() {
		for {
			health := checkServerHealth(ctx, healthClient)
			if health != nil {
				abnormal_counts++
				log.Printf(
					"%d / %d abnormal counts: status %v",
					abnormal_counts,
					abnormal_counts_limit,
					health,
				)
			} else {
				abnormal_counts = 0
			}

			if abnormal_counts <= abnormal_counts_limit {
				ping := &pb.HeartbeatPing{
					RegistrationId: registrationID,
					Timestamp:      time.Now().Unix(),
				}
				if err := stream.Send(ping); err != nil {
					log.Printf("Error sending ping: %v", err)
					return
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Receive heartbeats
	for {
		_, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving pong: %v", err)
			log.Printf("The manager has become unavailable")
			return
		}
	}
}

type FileWatcher struct {
	watcher    *fsnotify.Watcher
	debounce   time.Duration
	onModified func()
}

func NewFileWatcher(debounceTime time.Duration, onModified func()) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:    w,
		debounce:   debounceTime,
		onModified: onModified,
	}, nil
}

func (fw *FileWatcher) Watch(path string) error {
	if err := fw.watcher.Add(path); err != nil {
		return err
	}

	go fw.watchLoop()
	return nil
}

func (fw *FileWatcher) watchLoop() {
	debounce := time.Now()
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 && time.Since(debounce) > fw.debounce {
				fmt.Println("File changed, restarting...")
				fw.onModified()
				debounce = time.Now()
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Watcher error:", err)
		}
	}
}

func (fw *FileWatcher) Close() error {
	return fw.watcher.Close()
}
