package sshclient

import (
	"context"
	"fmt"
	"log"

	"github.com/xecus/connectedcar/config"
	"github.com/xecus/connectedcar/tunnel"
	"github.com/xecus/connectedcar/tunnel/util"
)

// local service to be forwarded
var localEndpoint = tunnel.PortFowardEndpoint{
	Host: "localhost",
	Port: 8090,
}

// remote SSH server
var serverEndpoint = tunnel.SSHServerEndpoint{
	Host: "localhost",
	Port: 2222,
}

// remote forwarding port (on remote SSH server network)
var remoteEndpoint = tunnel.PortFowardEndpoint{
	Host: "localhost",
	Port: 8080,
}

func SshConnectionClient(globalConfig *config.GlobalConfig, ctx context.Context) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tunnelConfig := &tunnel.TunnelConfig{
		SshServerEndpoint: &tunnel.SSHServerEndpoint{
			Host: "localhost",
			Port: 2222,
		},
		SshClientConfig: &tunnel.SSHClientConfig{
			User:          "admin+xxx",
			PublicKeyPath: "/home/hiroyuki/id_rsa",
		},
		LocalToRemoteForwarder: []*tunnel.PortfowardSrcDst{},
		RemoteToLocalForwarder: []*tunnel.PortfowardSrcDst{
			// 8080 -> 8090
			&tunnel.PortfowardSrcDst{
				Src: &tunnel.PortFowardEndpoint{
					Host: "localhost",
					Port: 8080,
				},
				Dst: &tunnel.PortFowardEndpoint{
					Host: "localhost",
					Port: 8090,
				},
			},
			// 8081 -> 8090
			&tunnel.PortfowardSrcDst{
				Src: &tunnel.PortFowardEndpoint{
					Host: "localhost",
					Port: 8081,
				},
				Dst: &tunnel.PortFowardEndpoint{
					Host: "localhost",
					Port: 22,
				},
			},
		},
	}

	sshClient, err := util.NewSSHClient(tunnelConfig)
	if err != nil {
		panic("Failed to Init SSH Connection")
	}
	defer sshClient.Close()
	go sshClient.KeepAlive(ctx)

	// Listen on remote server port
	err = sshClient.ListenPortOnRemote()
	if err != nil {
		panic("Failed to Listen Remote Port")
	}
	for _, listener := range sshClient.GetRemoteListeners() {
		defer func() {
			log.Printf("Called.")
			listener.Close()
		}()
	}
	go sshClient.BridgeLocalAndRemoteConnection(ctx)

	select {
	case <-ctx.Done():
		fmt.Println(ctx.Err())
	}

}
