package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func LoadConfig(configName string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType("toml")

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	v.AddConfigPath(currentDir)
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("config file was not found, make sure to include a %s file with your configuration", configName))
		} else {
			panic(fmt.Errorf("could not read config file: %w", err))
		}
	}
	return v, nil
}

func GetDependencies(v *viper.Viper) []Service {
	deps := v.GetStringMap("dependencies")

	var dependencies []Service

	for name, version := range deps {
		if versionStr, ok := version.(string); ok {
			dependencies = append(dependencies, Service{Name: name, Version: versionStr})
		} else {
			panic(fmt.Errorf("invalid type for version: expected string but got %T", version))
		}
	}

	fmt.Println("Dependencies:", dependencies)
	return dependencies
}
