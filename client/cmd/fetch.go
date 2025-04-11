package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	pb "otter.client/pb"
)

var fetch = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch latest service stubs",
	Long:  "Fetch the stubs defined in your otter.client.toml configuration file",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := ctx.Value("client").(pb.ManagerClient)
		fmt.Println("Fetching stubs...")

		v, err := LoadConfig("otter.client.toml")
		if err != nil {
			panic(fmt.Errorf("could not load config file: %v", err))
		}
		language := v.GetString("configuration.language")
		dependencies := GetDependencies(v)

		GetStubs(client, ctx, dependencies, language)
	},
}

func init() {
	fetch.Flags()
	rootCmd.AddCommand(fetch)
}
