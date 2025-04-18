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
		client := ctx.Value("client").(pb.ManagerClient)

		v, err := LoadConfig("h2pcontrol.server.toml")
		if err != nil {
			panic(fmt.Errorf("could not load config file: %v", err))
		}
		runCommand := v.GetString("configuration.run")
		protoPath := v.GetString("configuration.proto")

		service := pb.ServiceDefinition{
			ServiceName: v.GetString("service.name"),
			Version:     v.GetString("service.version"),
		}

		Run(client, ctx, runCommand, service, protoPath)
	},
}

func init() {
	run.Flags()
	rootCmd.AddCommand(run)
}
