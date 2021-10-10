package sshclient

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/xecus/connectedcar/config"
	"github.com/xecus/connectedcar/tunnel"
	"github.com/xecus/connectedcar/tunnel/util"
)

func SshConnectionClient(globalConfig *config.GlobalConfig, ctx context.Context) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		panic("Could not get homeDir")
	}
	privateKeyPath := filepath.Join(homeDirPath, ".ssh", "id_rsa")

	//TODO: Check Pubkey Permission

	tunnelConfig := &tunnel.TunnelConfig{
		// FIXME
		SshServerEndpoint: &tunnel.SSHServerEndpoint{
			Host: "localhost",
			Port: 2222,
		},
		// FIXME
		SshClientConfig: &tunnel.SSHClientConfig{
			User:          "admin+xxx",
			PublicKeyPath: privateKeyPath,
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
