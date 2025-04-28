package cmd

import (
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
		protocPath := filepath.Join(ep.BinDir(), "protoc")

		// Example: generate code
		cmd := exec.Command(protocPath,
			"--python_betterproto2_out="+outDir,
			"-I", protoDir,
			filepath.Join(protoDir, "*.proto"),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	},
}

func init() {
	compile.Flags()
	rootCmd.AddCommand(compile)
}
