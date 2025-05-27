package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
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
	pb "h2pcontrol.client/pb"
)

func Run(c pb.ManagerClient, ctx context.Context, runCommand string, server *pb.ServerDefinition, proto_path string) {

	// Check if the last entry is a file / check its path, then if it is a file and we have 1 arg it should just run that because
	// we can assume it's an executable? We can also check if it's an executable?
	args := strings.Fields(runCommand)
	if len(args) < 2 {
		fmt.Printf("invalid command format %s: need 'shell command'", args)
		return
	}
	filepath := args[len(args)-1]

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

	if err := fileWatcher.Watch(filepath); err != nil {
		panic(err)
	}

	// Register the server after it has been started, as otherwise it might not register even though it is not available
	// Do note, we are not checking if the python server itself has any errors.
	RegisterService(c, ctx, server, proto_path)

	go runHeartbeat(c)

	waitForShutdown(cmd)
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
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Server is running. Press Ctrl+C to stop.")

	sig := <-signalChan
	log.Printf("Received signal: %s. Shutting down...\n", sig)

	cmd.Process.Signal(os.Interrupt)

}

func RegisterService(c pb.ManagerClient, ctx context.Context, server *pb.ServerDefinition, proto_dir_path string) {

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
	r, err := c.RegisterServer(ctx, &request)
	if err != nil {
		log.Fatalf("Unable to connect to h2pcontrol Manager, is it running? %v", err)
	}
	log.Printf("RegisterServer call took %v\n", time.Since(rpcStart))
	log.Println(r.Result)

}

func runHeartbeat(client pb.ManagerClient) {
	// This is a bidirectional streaming heartbeat.
	stream, err := client.Heartbeat(context.Background())
	if err != nil {
		log.Fatalf("Failed to start heartbeat: %v", err)
	}

	// Send heartbeats
	go func() {
		for {
			ping := &pb.HeartbeatPing{
				Timestamp: time.Now().Unix(),
			}
			if err := stream.Send(ping); err != nil {
				log.Printf("Error sending ping: %v", err)
				return
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
		// log.Printf("Received heartbeat from manager: healthy=%v, ts=%d", pong.Healthy, pong.Timestamp)
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
