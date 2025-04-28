package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kluctl/go-embed-python/embed_util"
	"h2pcontrol.client/internal/assets/python-libs/data"

	"github.com/kluctl/go-embed-python/python"
	"github.com/spf13/cobra"
	"h2pcontrol.client/internal"
)

const colorPurple = "\033[35m"
const colorNone = "\033[0m"

// todo: make into nice functions :)

// should:
// build
// put in site-packages
// error handling
// nice printing

var compile = &cobra.Command{
	Use:   "compile",
	Short: "Compile proto files",
	Long:  "Compile your proto files and put them in your site-packages",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		protoDir := filepath.Join(args[0])

		ep, err := python.NewEmbeddedPython("betterproto")
		check(err)

		tmpDir := filepath.Join(os.TempDir(), "go-h2pcontrol")
		// Get the python libraries
		ef, err := embed_util.NewEmbeddedFilesWithTmpDir(data.Data, tmpDir+"-h2pcontrol-lib", true)
		check(err)

		extractedPath := ef.GetExtractedPath()
		ep.AddPythonPath(extractedPath)

		protocPath, err := internal.ExtractProtoc()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract protoc: %v\n", err)
			os.Exit(1)
		}

		outDir := filepath.Join("test", "tmp")
		os.MkdirAll(outDir, 0755)

		// Compile proto files
		fmt.Printf("%s[1/4] Compiling proto files...\n%s", colorPurple, colorNone)
		err = protoCompile(protocPath, protoDir, outDir)
		check(err)

		// Move the pyproject.toml file, required to make a build
		fmt.Printf("%s[2/4] Copying pyproject.toml...\n%s", colorPurple, colorNone)
		copyPyrojectToml(outDir)

		// Build the package
		fmt.Printf("%s[3/4] Building package...\n%s", colorPurple, colorNone)
		err = buildPackage(ep, outDir)
		check(err)

		// Install the dist locally
		fmt.Printf("%s[4/4] Installing package locally...\n%s", colorPurple, colorNone)

		distFile, err := findBuildPackage(outDir)
		check(err)
		installPackage(distFile)

		fmt.Printf("%sPackage installed into your Python environment's site-packages.%s\n", colorPurple, colorNone)
	},
}

func installPackage(distFile string) {
	pipCmd := exec.Command("python3", "-m", "pip", "install", distFile)
	// pipCmd.Stdout = os.Stdout
	pipCmd.Stderr = os.Stderr
	if err := pipCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pip install failed: %v\n", err)
		os.Exit(1)
	}
}

func findBuildPackage(outDir string) (string, error) {
	distDir := filepath.Join(outDir, "dist")
	files, err := os.ReadDir(distDir)
	if err != nil {
		panic(err)
	}

	var distFile string
	for _, f := range files {
		if !f.IsDir() && (filepath.Ext(f.Name()) == ".gz" || filepath.Ext(f.Name()) == ".whl") {
			distFile = filepath.Join(distDir, f.Name())
			break
		}
	}
	if distFile == "" {
		return "", fmt.Errorf("no .tar.gz or .whl file found in dist directory")
	}
	return distFile, nil
}

func copyPyrojectToml(outDir string) {
	srcFile, err := os.Open("pyproject.toml")
	check(err)
	defer srcFile.Close()
	destFile, err := os.Create(filepath.Join(outDir, "pyproject.toml"))
	check(err)
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	check(err)

	err = destFile.Sync()
	check(err)
}

func protoCompile(protocPath string, protoDir string, outDir string) error {

	protoFiles, err := filepath.Glob(filepath.Join(protoDir, "*.proto"))
	if err != nil || len(protoFiles) == 0 {
		return fmt.Errorf("no .proto files found in %s", protoDir)
	}

	protocArgs := []string{
		"--python_betterproto2_out=" + outDir,
		"-I", protoDir,
	}
	protocArgs = append(protocArgs, protoFiles...)
	protocCmd := exec.Command(protocPath, protocArgs...)

	protocCmd.Stdout = os.Stdout
	protocCmd.Stderr = os.Stderr
	if err := protocCmd.Run(); err != nil {
		return fmt.Errorf("failed to run protoc command: %v", err)
	}
	return nil
}

func buildPackage(ep *python.EmbeddedPython, outDir string) error {
	buildCmd, err := ep.PythonCmd("-m", "build", "--no-isolation", outDir)
	if err != nil {
		panic(err)
	}
	// buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("python build failed: %v", err)
	}
	return nil
}

func check(err error) {
	if err != nil {
		fmt.Printf("Error : %s\n", err.Error())
		os.Exit(1)
	}
}

func init() {
	compile.Flags()
	rootCmd.AddCommand(compile)
}
