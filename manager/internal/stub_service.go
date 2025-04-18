package internal

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	pb "h2pcontrol.manager/pb"
)

type StubService struct {
}

func NewStubService() *StubService {
	return &StubService{}
}

func (r *StubService) GetStub(_ context.Context, in *pb.StubRequest) (*pb.StubResponse, error) {
	log.Printf("Received call for service: '%v %v' for '%v'", in.GetServerName(), in.GetVersion(), in.GetLanguage())

	proto_path := filepath.Join("proto", in.GetServerName(), in.GetVersion())
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
