package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "h2pcontrol.client/pb"
)

var rootCmd = &cobra.Command{
	Use: "h2pcontrol",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize gRPC connection
		conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("connection failed: %v", err)
		}

		// Store connection and context in command
		ctx := context.Background()
		cmd.SetContext(context.WithValue(ctx, "conn", conn))
		cmd.SetContext(context.WithValue(cmd.Context(), "client", pb.NewManagerClient(conn)))
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Cleanup connection
		if conn, ok := cmd.Context().Value("conn").(*grpc.ClientConn); ok && conn != nil {
			conn.Close()
		}
	},
	Short: "h2pcontrol is a tool for managing grpc communication between different services",
	Long:  "h2pcontrol is a tool for managing grpc communication between different services. This is the h2pcontrol client which allows you to register your service and consume other services. ",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// Can also read these flags out of the config file.
// func init() {
// 	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", "localhost:50051", "Server address")
// 	rootCmd.PersistentFlags().StringVar(&service_name, "service_name", "arduino", "Service name")
// 	// ... other flag bindings
// }

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing Zero '%s'\n", err)
		os.Exit(1)
	}
}
