package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	pb "h2pcontrol.client/pb"
)

var fetch = &cobra.Command{
	Use:   "fetch [addr]",
	Short: "Fetch active servers or details about a specific server",
	Long:  "Fetch all active servers, or fetch details about a specific server if an address is provided.",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := ctx.Value("client").(pb.ManagerClient)

		if len(args) == 0 {
			// No argument: fetch all servers
			r, err := client.FetchServers(ctx, &pb.Empty{})
			if err != nil {
				log.Fatalf("could not fetch servers: %v", err)
			}
			PrettyPrintServers(r)
		} else {
			// Argument provided: fetch specific server
			r, err := client.FetchSpecificServer(ctx, &pb.FetchSpecificServerRequest{Addr: args[0]})
			if err != nil {
				log.Fatalf("could not fetch server: %v", err)
			}
			PrettyPrintServer(r)
		}
	},
}

func init() {
	rootCmd.AddCommand(fetch)
}

func PrettyPrintServers(resp *pb.FetchServersResponse) {
	if len(resp.Servers) == 0 {
		fmt.Println("No servers found.")
		return
	}
	fmt.Println("Registered Servers:")
	for _, server := range resp.Servers {
		fmt.Printf("  %s:\n", server.Name)
		fmt.Printf("    Description: %s\n", server.GetDescription())
		fmt.Printf("    Addr       : %s\n", server.GetAddr())
		fmt.Println()
	}
}

func PrettyPrintServer(resp *pb.FetchSpecificServerResponse) {
	fmt.Printf("  %s:\n", resp.ServerDefinition.Name)
	fmt.Printf("    Description: %s\n", resp.ServerDefinition.GetDescription())
	fmt.Printf("    Addr       : %s\n", resp.ServerDefinition.GetAddr())
	fmt.Printf("    Proto      : \n\n%s\n", resp.GetProto())
	fmt.Println()
}
