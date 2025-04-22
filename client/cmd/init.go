package cmd

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	templates "h2pcontrol.client/internal/templates"
)

var init_cmd = &cobra.Command{
	Use:   "init [server/client]",
	Short: "Initialize an h2pcontrol server or client in the current directory.",
	Long:  "Initialize an h2pcontrol server or client in the current directory.",
	Run: func(cmd *cobra.Command, args []string) {
		projectType := promptProjectType()
		if projectType == "" {
			return
		}

		projectName := promptProjectName()
		if projectName == "" {
			return
		}

		fmt.Printf("You chose to create a %s named %q\n", projectType, projectName)

		if err := copyEmbedDir(templates.TemplatesFS, projectType, "."); err != nil {
			fmt.Printf("Failed to copy template: %v\n", err)
			return
		}

		// Rename files and replace placeholders
		if err := renameAndReplace(projectType, projectName); err != nil {
			fmt.Printf("Failed to finalize project: %v\n", err)
			return
		}

		fmt.Printf("Succesfully initialized %s\n", projectName)
	},
}

func renameAndReplace(projectType, projectName string) error {
	// Rename folders/files and replace placeholders
	if projectType == "server" {
		// Rename files/folders
		if err := os.Rename("src/_name_", "src/"+projectName); err != nil {
			return err
		}

		if err := os.Rename("proto/_name_.proto", "proto/"+projectName+".proto"); err != nil {
			return err
		}
		// Rename __name__ in h2pcontrol.server.toml, pyproject.toml and in proto.toml
		replacePlaceholderValues("pyproject.toml", "_name_", projectName)
		replacePlaceholderValues("h2pcontrol.server.toml", "_name_", projectName)
		replacePlaceholderValues(filepath.Join("proto", projectName+".proto"), "_name_", projectName)
	} else if projectType == "client" {

		// replacePlaceholderValues("pyproject.toml", "_name_", projectName)
	}
	return nil
}

func promptProjectType() string {
	prompt := promptui.Select{
		Label: "Do you want to create a server or client?",
		Items: []string{"client", "server"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}
	return result
}

func promptProjectName() string {
	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("name cannot be empty")
		}
		match, err := regexp.MatchString("^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9_.-]*[a-zA-Z0-9])$", input)
		if !match || err != nil {
			return errors.New("name must be ASCII letters/numbers, period, underscore, hyphen; must start/end with letter/number")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Project name",
		Validate: validate,
	}
	name, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}
	return name
}

// Try to recreate the otiai10 copy functionality somewhat. This is necessary as we are using the embed FileSystem,
// as we will embed the client/server templates inside of our binary.
func copyEmbedDir(efs embed.FS, src, dst string) error {
	entries, err := efs.ReadDir(src)
	if err != nil {
		fmt.Printf("Could not read source dir")
		return err
	}
	// https://chmodcommand.com/chmod-0755/
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyEmbedDir(efs, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := efs.ReadFile(srcPath)
			if err != nil {
				return err
			}
			// https://chmodcommand.com/chmod-0644/
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func replacePlaceholderValues(filePath string, placeholder string, value string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := strings.ReplaceAll(string(data), placeholder, value)
	// https://chmodcommand.com/chmod-0644/
	os.WriteFile(filePath, []byte(content), 0644)
	return nil
}

func init() {
	rootCmd.AddCommand(init_cmd)
}
