package cmd

import (
	"errors"
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

const (
	colorPurple = "\033[37m"
	colorNone   = "\033[0m"

	pythonBetterprotoPackage = "betterproto"
	pythonBuildModule        = "build"
	pythonPipCommand         = "pip"
	pyprojectTomlFile        = "pyproject.toml"

	tempDirPrefix = "go-h2pcontrol"
	tempLibSuffix = "-h2pcontrol-lib"
	tempBuildDir  = "h2pcontrol-temp"
	srcDirName    = "src"
	distDirName   = "dist"

	wheelExt = ".whl"
	gzipExt  = ".gz"

	buildNoIsolationFlag = "--no-isolation"
	pipForceReinstall    = "--force-reinstall"

	protocPythonOutFlag = "--python_betterproto2_out="
	protocIncludeFlag   = "-I"
	protoFilePattern    = "*.proto"
)

var progressMessages = [...]string{
	"[1/4] Compiling proto files...",
	"[2/4] Copying pyproject.toml...",
	"[3/4] Building package...",
	"[4/4] Installing package locally...",
}

const (
	errProtoDirRequired = "Error: --proto-dir is required"
	errNameRequired     = "Error: --name is required"
	errNoProtoFiles     = "no .proto files found in %s"
	errProtocFailed     = "failed to run protoc command: %v"
	errPythonBuild      = "python build failed: %v"
	errPipInstall       = "pip install failed: %v"
	errNoDistFile       = "no .tar.gz or .whl file found in dist directory"
	errExtractProtoc    = "Failed to extract protoc: %v"
)

var (
	protoDir string
	name     string
)

var compile = &cobra.Command{
	Use:   "compile",
	Short: "Compile proto files",
	Long:  "Compile your proto files and put them in your site-packages",
	Run: func(cmd *cobra.Command, args []string) {
		if protoDir == "" {
			fmt.Fprintf(os.Stderr, "Error: --proto-dir is required\n")
			os.Exit(1)
		}
		if name == "" {
			fmt.Fprintf(os.Stderr, "Error: --name is required\n")
			os.Exit(1)
		}

		ep, err := python.NewEmbeddedPython("betterproto")
		check(err)

		tmpDir := filepath.Join(os.TempDir(), "go-h2pcontrol")

		// Get the python libraries we embedded
		ef, err := embed_util.NewEmbeddedFilesWithTmpDir(data.Data, tmpDir+"-h2pcontrol-lib", true)
		check(err)

		extractedPath := ef.GetExtractedPath()
		ep.AddPythonPath(extractedPath)

		protocPath, err := internal.ExtractProtoc()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract protoc: %v\n", err)
			os.Exit(1)
		}

		outDir := filepath.Join(os.TempDir(), "h2pcontrol-temp")
		os.MkdirAll(outDir, 0755)

		// Compile proto files
		fmt.Printf("%s%s\n%s", colorPurple, progressMessages[0], colorNone)
		protoOutDir := outDir + fmt.Sprintf("/src/%s/", name)
		os.MkdirAll(protoOutDir, 0755)
		err = protoCompile(protocPath, protoDir, protoOutDir)
		check(err)

		// Move the pyproject.toml file, required to make a build
		fmt.Printf("%s%s\n%s", colorPurple, progressMessages[1], colorNone)
		copyPyrojectToml(outDir)

		// Build the package
		fmt.Printf("%s%s\n%s", colorPurple, progressMessages[2], colorNone)
		err = buildPackage(ep, outDir)
		check(err)

		// Install the dist locally
		fmt.Printf("%s%s\n%s", colorPurple, progressMessages[3], colorNone)

		distFile, err := findBuildPackage(outDir)
		check(err)
		installPackage(distFile)

		fmt.Printf("%sPackage installed into your Python environment's site-packages.%s\n", colorPurple, colorNone)

		// Clean up
		err = os.RemoveAll(outDir)
		check(err)
	},
}

func installPackage(distFile string) {
	pipCmd := exec.Command(pythonPipCommand, "install", pipForceReinstall, distFile)
	pipCmd.Stdout = os.Stdout
	pipCmd.Stderr = os.Stderr
	if err := pipCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, errPipInstall+"\n", err)
		os.Exit(1)
	}
}

func findBuildPackage(outDir string) (string, error) {
	distDir := filepath.Join(outDir, distDirName)
	files, err := os.ReadDir(distDir)
	if err != nil {
		panic(err)
	}

	var distFile string
	for _, f := range files {
		if !f.IsDir() && (filepath.Ext(f.Name()) == gzipExt || filepath.Ext(f.Name()) == wheelExt) {
			distFile = filepath.Join(distDir, f.Name())
			break
		}
	}
	if distFile == "" {
		return "", errors.New(errNoDistFile)
	}
	return distFile, nil
}

func copyPyrojectToml(outDir string) {
	srcFile, err := os.Open(pyprojectTomlFile)
	check(err)
	content, err := io.ReadAll(srcFile)
	check(err)
	fmt.Printf("Contents of %s:\n%s\n", pyprojectTomlFile, string(content))

	// Rewind the file pointer to the beginning for copying
	_, err = srcFile.Seek(0, io.SeekStart)
	check(err)
	defer srcFile.Close()
	destFile, err := os.Create(filepath.Join(outDir, pyprojectTomlFile))
	check(err)
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	check(err)

	err = destFile.Sync()
	check(err)
}

func protoCompile(protocPath string, protoDir string, outDir string) error {

	protoFiles, err := filepath.Glob(filepath.Join(protoDir, protoFilePattern))
	if err != nil || len(protoFiles) == 0 {
		return fmt.Errorf(errNoProtoFiles, protoDir)
	}

	protocArgs := []string{
		protocPythonOutFlag + outDir,
		protocIncludeFlag, protoDir,
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
	buildCmd.Stdout = os.Stdout
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
	compile.Flags().StringVar(&protoDir, "proto-dir", "", "Directory containing proto files (required)")
	compile.Flags().StringVar(&name, "name", "", "Name of the package (required)")

	compile.MarkFlagRequired("proto-dir")
	compile.MarkFlagRequired("name")

	rootCmd.AddCommand(compile)
}
