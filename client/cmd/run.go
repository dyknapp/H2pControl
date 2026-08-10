package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	pb "h2pcontrol.client/pb"
)

var run = &cobra.Command{
	Use:   "run",
	Short: "Run server",
	Long:  "Run your server, connect to the manager and make your server available for others to call.",
	Args:  cobra.ArbitraryArgs,
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

		advertiseSubnet := h2p_config.GetString(
			"configuration.advertise_subnet",
		)

		if advertiseSubnet == "" {
			panic(fmt.Errorf(
				"configuration.advertise_subnet is missing from %s",
				h2p_config.ConfigFileUsed(),
			))
		}

		fmt.Printf("Loaded configuration from: %s\n", h2p_config.ConfigFileUsed())
		fmt.Printf("Advertise subnet: %q\n", advertiseSubnet)

		advertisedHost, err := getIPInSubnet(advertiseSubnet)
		if err != nil {
			panic(fmt.Errorf("Could not determine the advertised host: %w", err))
		}

		serverName := h2p_config.GetString("configuration.server_name")
		if serverName == "" {
			serverName = pyproject_config.GetString("project.name")
			log.Printf(
				"configuration.server_name is missing; using project.name %q",
				serverName,
			)
		}

		service := pb.ServerDefinition{
			AdvertisedHost: advertisedHost,
			Port:           h2p_config.GetString("configuration.port"),
			ServerName:     serverName,
			Version:        pyproject_config.GetString("project.version"),
		}

		Run(client, ctx, runCommand, &service, protoPath, args)
	},
}

func init() {
	rootCmd.AddCommand(run)
}
