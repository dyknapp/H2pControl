package cmd

import (
	"github.com/spf13/cobra"
	pb "otter.client/pb"
)

var run = &cobra.Command{
	Use:   "run",
	Short: "Run server",
	Long:  "Run your server, connect to the manager and make your server available for others to call.",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := ctx.Value("client").(pb.ManagerClient)

		
		Run(client, ctx)
	},
}

func init() {
	run.Flags()
	rootCmd.AddCommand(run)
}
