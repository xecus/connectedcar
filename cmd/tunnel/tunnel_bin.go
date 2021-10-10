package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/xecus/connectedcar/cmd/tunnel/sshclient"
	"github.com/xecus/connectedcar/cmd/tunnel/sshserver"
	"github.com/xecus/connectedcar/config"
)

func main() {
	flag.Parse()

	globalConfig, err := config.NewConfig()
	if err != nil {
		panic("Failed to init config.")
	}

	switch flag.Arg(0) {
	case "server":
		log.Println("Server")
		sshserver.SshdWithPortForwarding(globalConfig)
	case "client":

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		defer stop()

		log.Println("Client")
		sshclient.SshConnectionClient(globalConfig, ctx)
	default:
		log.Println("Default")
	}

}
