package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/python"
	"github.com/spf13/cobra"
	"h2pcontrol.client/internal"
)

var compile = &cobra.Command{
	Use:   "compile",
	Short: "Compile proto files",
	Long:  "Compile your proto files and put them in your site-packages",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		protoDir := filepath.Join(args[0])

		ep, err := python.NewEmbeddedPython("betterproto")
		if err != nil {
			return
		}

		ef, err := embed_util.NewEmbeddedFiles(&internal.PythonLibs, "internal/assets/python-libs/data")
		extractedPath := ef.GetExtractedPath()
		if err != nil {
			return
		}
		ep.AddPythonPath(extractedPath)

		// test:
		protocPath, err := internal.ExtractProtoc()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract protoc: %v\n", err)
			os.Exit(1)
		}

		outDir := filepath.Join("test", "tmp")

		protocCmd := exec.Command(protocPath,
			"--python_betterproto2_out="+outDir,
			"-I", protoDir,
			filepath.Join(protoDir, "*.proto"),
		)
		protocCmd.Stdout = os.Stdout
		protocCmd.Stderr = os.Stderr
		if err := protocCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to run protoc command: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	compile.Flags()
	rootCmd.AddCommand(compile)
}
