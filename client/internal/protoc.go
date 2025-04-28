package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// BINARIES FROM https://github.com/protocolbuffers/protobuf/releases
func ExtractProtoc() (string, error) {
	var dir, binName string

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			dir = "protoc-25.7-linux-x86_64"
		case "386":
			dir = "protoc-25.7-linux-x86_32"
		case "arm64":
			dir = "protoc-25.7-linux-aarch_64"
		case "ppc64le":
			dir = "protoc-25.7-linux-ppcle_64"
		case "s390x":
			dir = "protoc-25.7-linux-s390_64"
		default:
			return "", fmt.Errorf("unsupported Linux arch: %s", runtime.GOARCH)
		}
		binName = "protoc"
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			dir = "protoc-25.7-osx-x86_64"
		case "arm64":
			dir = "protoc-25.7-osx-aarch_64"
		default:
			// Universal binary fallback
			dir = "protoc-25.7-osx-universal_binary"
		}
		binName = "protoc"
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			dir = "protoc-25.7-win64"
		case "386":
			dir = "protoc-25.7-win32"
		default:
			return "", fmt.Errorf("unsupported Windows arch: %s", runtime.GOARCH)
		}
		binName = "protoc.exe"
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	assetPath := filepath.Join("internal/assets/protoc", dir, "bin", binName)
	binData, err := ProtocBinaries.Open(assetPath)
	if err != nil {
		return "", fmt.Errorf("failed to open embedded protoc: %w", err)
	}
	defer binData.Close()

	tmpDir, err := os.MkdirTemp("", "protoc-bin")
	if err != nil {
		return "", err
	}
	outPath := filepath.Join(tmpDir, binName)
	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, binData); err != nil {
		return "", err
	}
	if !strings.HasSuffix(binName, ".exe") {
		os.Chmod(outPath, 0755)
	}
	return outPath, nil
}
