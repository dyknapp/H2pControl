package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	pb "h2pcontrol.client/pb"
)

var stubCmd = &cobra.Command{
	Use:   "stub",
	Short: "Retrieve and install Python server stubs",
	Long:  "Servers transfer their proto definitions to the manager when they are registered.  The manager will generate a stub and produce a .whl file that will be transferred to your workspace so that you can call the server using gRPC.",
}

var stubGetCmd = &cobra.Command{
	Use:   "get SERVER",
	Short: "Retrieve and install generated server stub",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, ok := ctx.Value(clientKey).(pb.ManagerClient)
		if !ok || client == nil {
			return fmt.Errorf("manager client is not initialized")
		}

		response, err := client.GetStub(ctx, &pb.StubRequest{
			ServerName: args[0],
			Version:    stubVersion,
			Language:   "python",
		})
		if err != nil {
			return fmt.Errorf(
				"get stub for %s version %s: %w",
				args[0],
				stubVersion,
				err,
			)
		}

		path, err := writeStubWheel(".", response)
		if err != nil {
			return err
		}

		fmt.Fprintf(
			cmd.OutOrStdout(),
			"Wrote %s (SHA-256: %s)\n",
			path,
			response.GetWheelSha256(),
		)
		return nil
	},
}

var stubInstallCmd = &cobra.Command{
	Use:   "install WHEEL",
	Short: "Install a generated server stub in the local uv project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wheelPath, projectDir, err := installStubWheel(
			cmd.Context(),
			".",
			args[0],
			cmd.OutOrStdout(),
			cmd.ErrOrStderr(),
		)
		if err != nil {
			return err
		}

		fmt.Fprintf(
			cmd.OutOrStdout(),
			"Installed %s in uv project %s\n",
			filepath.Base(wheelPath),
			projectDir,
		)
		return nil
	},
}

var stubVersion string

func writeStubWheel(outputDir string, response *pb.StubResponse) (string, error) {
	if response == nil {
		return "", fmt.Errorf("manager returned an empty stub response")
	}

	filename := response.GetFilename()
	if filename == "" || filepath.Base(filename) != filename ||
		strings.ContainsAny(filename, `/\`) ||
		!strings.EqualFold(filepath.Ext(filename), wheelExt) {
		return "", fmt.Errorf("manager returned invalid wheel filename %q", filename)
	}

	expectedSHA256, err := hex.DecodeString(response.GetWheelSha256())
	if err != nil || len(expectedSHA256) != sha256.Size {
		return "", fmt.Errorf(
			"manager returned invalid wheel SHA-256 %q",
			response.GetWheelSha256(),
		)
	}

	actualSHA256 := sha256.Sum256(response.GetWheelData())
	if !bytes.Equal(actualSHA256[:], expectedSHA256) {
		return "", fmt.Errorf(
			"wheel checksum mismatch: expected %s, got %x",
			response.GetWheelSha256(),
			actualSHA256,
		)
	}

	path := filepath.Join(outputDir, filename)
	if err := os.WriteFile(path, response.GetWheelData(), 0o644); err != nil {
		return "", fmt.Errorf("write stub wheel %q: %w", path, err)
	}

	return path, nil
}

func installStubWheel(
	ctx context.Context,
	projectDir string,
	wheelPath string,
	stdout io.Writer,
	stderr io.Writer,
) (string, string, error) {
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve uv project directory: %w", err)
	}

	projectFile := filepath.Join(projectDir, pyprojectTomlFile)
	projectInfo, err := os.Stat(projectFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf(
				"no %s found in uv project directory %q",
				pyprojectTomlFile,
				projectDir,
			)
		}
		return "", "", fmt.Errorf("inspect uv project file %q: %w", projectFile, err)
	}
	if !projectInfo.Mode().IsRegular() {
		return "", "", fmt.Errorf("uv project file %q is not a regular file", projectFile)
	}

	if !filepath.IsAbs(wheelPath) {
		wheelPath = filepath.Join(projectDir, wheelPath)
	}
	wheelPath, err = filepath.Abs(wheelPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve stub wheel path: %w", err)
	}
	if !strings.EqualFold(filepath.Ext(wheelPath), wheelExt) {
		return "", "", fmt.Errorf("stub package %q is not a .whl file", wheelPath)
	}
	wheelInfo, err := os.Stat(wheelPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("stub wheel %q does not exist", wheelPath)
		}
		return "", "", fmt.Errorf("inspect stub wheel %q: %w", wheelPath, err)
	}
	if !wheelInfo.Mode().IsRegular() {
		return "", "", fmt.Errorf("stub wheel %q is not a regular file", wheelPath)
	}

	uvCmd := exec.CommandContext(ctx, "uv", "add", wheelPath)
	uvCmd.Dir = projectDir
	uvCmd.Stdout = stdout
	uvCmd.Stderr = stderr
	if err := uvCmd.Run(); err != nil {
		return "", "", fmt.Errorf("install stub wheel with uv: %w", err)
	}

	return wheelPath, projectDir, nil
}

func init() {
	stubGetCmd.Flags().StringVar(
		&stubVersion,
		"version",
		"",
		"Exact server interface version",
	)

	stubGetCmd.MarkFlagRequired("version")

	stubCmd.AddCommand(stubGetCmd, stubInstallCmd)
	rootCmd.AddCommand(stubCmd)
}
