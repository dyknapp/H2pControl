package cmd

import (
	"archive/zip"
	"bufio"
	"bytes"
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
	"google.golang.org/grpc"
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
				cmd.Process.Signal(os.Interrupt) // Sends CTRL_BREAK_EVENT
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
	for {

		_, err := client.Heartbeat(context.Background(), &pb.Empty{}, grpc.EmptyCallOption{})
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

func GetStubs(c pb.ManagerClient, ctx context.Context, dependencies []pb.ServerDefinition, language string) {

	for _, dependency := range dependencies {
		GetStub(c, ctx, dependency.ServerName, dependency.Version, language)
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
	r, err := c.GetStub(ctx, &pb.StubRequest{ServerName: service_name, Version: version, Language: language})
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
