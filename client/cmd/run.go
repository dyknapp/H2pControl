package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	pb "h2pcontrol.client/pb"
)

var run = &cobra.Command{
	Use:   "run",
	Short: "Run server",
	Long:  "Run your server, connect to the manager and make your server available for others to call.",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client, ok := ctx.Value(clientKey).(pb.ManagerClient)
		if !ok || client == nil {
			panic("ManagerClient not initialized. Make sure you are connected to the manager first")
		}

		h2p_config, err := LoadConfig("h2pcontrol.server.toml")
		if err != nil {
			panic(fmt.Errorf("could not load h2pcontrol.server file: %v", err))
		}
		runCommand := h2p_config.GetString("configuration.run")
		protoPath := h2p_config.GetString("configuration.proto")

		pyproject_config, err := LoadConfig("pyproject.toml")
		if err != nil {
			panic(fmt.Errorf("could not load pyproject config file: %v", err))
		}


		service := pb.ServerDefinition{
			Port:       h2p_config.GetString("configuration.port"),
			ServerName: pyproject_config.GetString("project.name"),
			Version:    pyproject_config.GetString("project.version"),
		}

		Run(client, ctx, runCommand, &service, protoPath)
	},
}

func init() {
	run.Flags()
	rootCmd.AddCommand(run)
}


