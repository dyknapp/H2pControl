package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"h2pcontrol.client/internal"
)

const (
	colorPurple = "\033[37m"
	colorNone   = "\033[0m"

	protoPackagesDir = "proto_packages"

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
	errNoProtoFiles     = "no .proto files found in %s"
	errProtocFailed     = "failed to run protoc command: %v"
	errPythonBuild      = "uv build failed: %v"
	errPipInstall       = "uv pip install failed: %v"
	errNoDistFile       = "no .tar.gz or .whl file found in dist directory"
	errExtractProtoc    = "Failed to extract protoc: %v"
)

var (
	protoDir string
	stubName string
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

		h2p_config, err := LoadConfig("h2pcontrol.server.toml")
		if err != nil {
			panic(fmt.Errorf("could not load h2pcontrol.server file: %v", err))
		}
		if stubName == "" {
			serverName := h2p_config.GetString("configuration.server_name")
			if serverName == "" {
				panic(fmt.Errorf("could not parse h2pcontrol.server.toml server name: %v", err))
			}
			stubName = serverName + "-proto"
		}

		distributionName := normalizeDistributionName(stubName)
		moduleName := moduleNamefromDistribution(distributionName)

		packageDir := filepath.Join("proto_packages", moduleName)
		protoOutDir := filepath.Join(packageDir, "src", moduleName)

		// Get version from current project
		version, err := getPyProjectVersion()
		check(err)

		// Compile proto files directly into the package directory
		err = os.MkdirAll(protoOutDir, 0755)
		check(err)

		protocPath, err := internal.ExtractProtoc()
		check(err)

		fmt.Printf("%s%s\n%s", colorPurple, "Compiling proto files...", colorNone)
		err = protoCompile(protocPath, protoDir, protoOutDir)
		check(err)

		createPyProjectToml(packageDir, distributionName, version)

		installEditablePackage(packageDir)

		fmt.Printf("%sPackage installed. %s\n", colorPurple, colorNone)

	},
}

func getPyProjectVersion() (string, error) {
	srcFile, err := os.Open(pyprojectTomlFile)
	if err != nil {
		return "1.0", nil // Default version if file can't be read
	}
	defer srcFile.Close()

	content, err := io.ReadAll(srcFile)
	if err != nil {
		return "1.0", nil
	}

	var config map[string]interface{}
	err = toml.Unmarshal(content, &config)
	if err != nil {
		return "1.0", nil
	}

	if project, ok := config["project"].(map[string]interface{}); ok {
		if version, ok := project["version"].(string); ok {
			return version, nil
		}
	}
	return "1.0", nil
}

func createPyProjectToml(packageDir string, name string, version string) {
	config := map[string]interface{}{
		"build-system": map[string]interface{}{
			"requires":      []string{"setuptools", "wheel"},
			"build-backend": "setuptools.build_meta",
		},
		"project": map[string]interface{}{
			"name":            name,
			"version":         version,
			"description":     "Generated proto package",
			"requires-python": ">=3.11",
			"dependencies": []string{
				"betterproto2>=0.4.0",
			},
		},
		"tool": map[string]interface{}{
			"setuptools": map[string]interface{}{
				"package-dir": map[string]interface{}{
					"": "src",
				},
				"packages": map[string]interface{}{
					"find": map[string]interface{}{
						"where": []string{"src"},
					},
				},
			},
		},
	}

	content, err := toml.Marshal(config)
	check(err)

	err = os.WriteFile(filepath.Join(packageDir, pyprojectTomlFile), content, 0644)
	check(err)
}

func protoCompile(protocPath string, protoDir string, outDir string) error {
	protoFiles, err := filepath.Glob(filepath.Join(protoDir, protoFilePattern))
	if err != nil || len(protoFiles) == 0 {
		return fmt.Errorf(errNoProtoFiles, protoDir)
	}

	// Verify betterproto2_compiler is installed
	syncCmd := exec.Command("uv", "sync")
	syncCmd.Stdout = os.Stdout
	syncCmd.Stderr = os.Stderr

	if err := syncCmd.Run(); err != nil {
		return fmt.Errorf("uv sync failed: %w", err)
	}
	checkCmd := exec.Command("uv", "run", "python", "-c", "import betterproto2_compiler; print('betterproto2_compiler found')")
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

func installEditablePackage(distFile string) {
	addCmd := exec.Command("uv", "add", "--editable", distFile)
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

func normalizeDistributionName(name string) string {
	name = strings.ToLower(name)
	return regexp.MustCompile(`[-_.]+`).ReplaceAllString(name, "-")
}

func moduleNamefromDistribution(name string) string {
	return strings.ReplaceAll(normalizeDistributionName(name), "-", "_")
}

func init() {
	compile.Flags().StringVar(&protoDir, "proto-dir", "", "Directory containing proto files (required)")
	compile.Flags().StringVar(&stubName, "stubname", "", "Name of the package (optional override)")

	compile.MarkFlagRequired("proto-dir")

	rootCmd.AddCommand(compile)
}
