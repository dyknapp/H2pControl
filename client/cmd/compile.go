package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
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
	errPythonBuild      = "uv build failed: %v"
	errPipInstall       = "uv pip install failed: %v"
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

		pyProjectName, err := getPyProjectName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get pyproject name: %v\n", err)
			os.Exit(1)
		}
		if pyProjectName == name {
			fmt.Fprintf(os.Stderr, "The proto package can not have the same name as your python package, please use a different --name:\n")
			os.Exit(1)
		}

		protocPath, err := internal.ExtractProtoc()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract protoc: %v\n", err)
			os.Exit(1)
		}

		// TODO: fix this rm this
		outDir := filepath.Join(os.TempDir(), "h2pcontrol-temp")
		os.MkdirAll(outDir, 0755)

		// Compile proto files
		fmt.Printf("%s%s\n%s", colorPurple, progressMessages[0], colorNone)
		protoOutDir := outDir + fmt.Sprintf("/src/%s/", name)
		os.MkdirAll(protoOutDir, 0755)
		
		err = protoCompile(protocPath, protoDir, protoOutDir)
		check(err)

		fmt.Printf("Compiled files")

		// Move the pyproject.toml file, required to make a build
		fmt.Printf("%s%s\n%s", colorPurple, progressMessages[1], colorNone)
		copyPyProjectToml(outDir, name)

		// Build the package
		fmt.Printf("%s%s\n%s", colorPurple, progressMessages[2], colorNone)
		err = buildPackage(outDir)
		check(err)

		// Install the dist locally
		fmt.Printf("%s%s\n%s", colorPurple, progressMessages[3], colorNone)
		distFile, err := findBuildPackage(outDir)
		check(err)
		installPackage(distFile)

		fmt.Printf("%sPackage installed. %s\n", colorPurple, colorNone)

		// err = os.RemoveAll(outDir)
		// check(err)
	},
}

func findBuildPackage(outDir string) (string, error) {
	distDir := path.Join(outDir, distDirName)
	files, err := os.ReadDir(distDir)
	if err != nil {
		fmt.Printf("Error! %s", protoDir)
		panic(err)
	}

	var distFile string
	for _, f := range files {
		if !f.IsDir() && (path.Ext(f.Name()) == gzipExt || path.Ext(f.Name()) == wheelExt) {
			distFile = path.Join(distDir, f.Name())
			break
		}
	}
	if distFile == "" {
		return "", errors.New(errNoDistFile)
	}
	return distFile, nil
}

func getPyProjectName() (string, error) {
	srcFile, err := os.Open(pyprojectTomlFile)
	check(err)
	content, err := io.ReadAll(srcFile)
	check(err)

	var config map[string]interface{}
	err = toml.Unmarshal(content, &config)
	check(err)

	if project, ok := config["project"].(map[string]interface{}); ok {
		name, ok := project["name"]
		if ok {
			if nameStr, ok := name.(string); ok {
				return nameStr, nil
			}

		}
	}
	return "", errors.New("pyproject.toml could not be found")
}

func copyPyProjectToml(outDir string, name string) {
	srcFile, err := os.Open(pyprojectTomlFile)
	check(err)
	content, err := io.ReadAll(srcFile)
	check(err)

	// Set the name from the cmd line arguments
	var config map[string]interface{}
	err = toml.Unmarshal(content, &config)
	check(err)

	if project, ok := config["project"].(map[string]interface{}); ok {
		project["name"] = name

	}
	newContent, err := toml.Marshal(config)
	check(err)

	fmt.Printf("Modified contents of %s:\n%s\n", pyprojectTomlFile, string(newContent))

	destFile, err := os.Create(filepath.Join(outDir, pyprojectTomlFile))
	check(err)
	defer destFile.Close()

	_, err = destFile.WriteString(string(newContent))
	check(err)

	err = destFile.Sync()
	check(err)
}

func protoCompile(protocPath string, protoDir string, outDir string) error {
	protoFiles, err := filepath.Glob(filepath.Join(protoDir, protoFilePattern))
	if err != nil || len(protoFiles) == 0 {
		return fmt.Errorf(errNoProtoFiles, protoDir)
	}

	// Verify betterproto2_compiler is installed
	checkCmd := exec.Command("python", "-c", "import betterproto2_compiler; print('betterproto2_compiler found')")
	checkCmd.Stdout = os.Stdout
	checkCmd.Stderr = os.Stderr
	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf("betterproto2_compiler not found, please install it with: pip install betterproto2")
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

func buildPackage(outDir string) error {
	checkCmd := exec.Command("uv", "--version")
	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf("uv not found, please install it with: pip install uv")
	}

	buildCmd := exec.Command("uv", "build", outDir)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	return buildCmd.Run()
}

func installPackage(distFile string) {
	installCmd := exec.Command("uv", "pip", "install", distFile)
	installCmd.Dir = "."
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install package: %v\n", err)
		os.Exit(1)
	}

	// Add as dependency with --frozen flag
	addCmd := exec.Command("uv", "add", "--frozen", distFile)
	addCmd.Dir = "."
	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr
	if err := addCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, errPipInstall+"\n", err)
		os.Exit(1)
	}
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
