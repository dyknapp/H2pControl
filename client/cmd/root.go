package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "h2pcontrol.client/pb"
)

type contextKey string

const (
	connKey   contextKey = "conn"
	clientKey contextKey = "client"
)

var managerAddr string

var rootCmd = &cobra.Command{
	Use: "h2pcontrol",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		// Initialize gRPC connection
		conn, err := grpc.NewClient(managerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("connection failed: %v", err)
		}
		fmt.Printf("Connection established: %v\n", conn != nil)

		// Store connection and context in command
		ctx := context.Background()
		ctx = context.WithValue(ctx, connKey, conn)
		client := pb.NewManagerClient(conn)
		ctx = context.WithValue(ctx, clientKey, client)
		cmd.SetContext(ctx)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Cleanup connection
		if conn, ok := cmd.Context().Value(connKey).(*grpc.ClientConn); ok && conn != nil {
			conn.Close()
		}
	},
	Short: "h2pcontrol is a tool for managing grpc communication between different services",
	Long:  "h2pcontrol is a tool for managing grpc communication between different services. This is the h2pcontrol client which allows you to register your service and consume other services. ",
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&managerAddr,
		"manager",
		"127.0.0.1:50051",
		"Address of the manager gRPC server",
	)
	viper.BindPFlag("manager", rootCmd.PersistentFlags().Lookup("manager"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing Zero '%s'\n", err)
		os.Exit(1)
	}
}
