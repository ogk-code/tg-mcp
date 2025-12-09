package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
	"tg-mcp/tools"
)

func main() {
	cfg, err := client.ConfigFromEnv()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	tgClient := client.New(cfg)

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "tg-mcp",
			Version: "1.0.0",
		},
		nil,
	)

	tools.RegisterAuthTools(server, tgClient)
	tools.RegisterSendTools(server, tgClient)
	tools.RegisterMessagesTools(server, tgClient)
	tools.RegisterChatsTools(server, tgClient)
	tools.RegisterManageTools(server, tgClient)
	tools.RegisterUsersTools(server, tgClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- tgClient.Run(ctx, func(ctx context.Context) error {
			return server.Run(ctx, &mcp.StdioTransport{})
		})
	}()

	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			log.Fatalf("Error: %v", err)
		}
	case <-ctx.Done():
	}
}
