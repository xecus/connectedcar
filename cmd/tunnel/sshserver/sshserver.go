package sshserver

import (
	"io"
	"log"

	"github.com/gliderlabs/ssh"
	"github.com/xecus/connectedcar/config"
	originalssh "github.com/xecus/connectedcar/tunnel/ssh"
	"github.com/xecus/connectedcar/tunnel/util"
)

func SshdWithPortForwarding(globalConfig *config.GlobalConfig) {

	log.Println("Starting ssh server on port 2222...")

	forwardHandler := &originalssh.ForwardedTCPHandler{}

	server := ssh.Server{
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			log.Println("Accepted forward", dhost, dport)
			return true
		}),
		Addr: ":2222",
		Handler: ssh.Handler(func(s ssh.Session) {
			io.WriteString(s, "Remote forwarding available...\n")
			util.GenerateSessionHandler(globalConfig)(s)
			//select {}
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			log.Println("attempt to bind", host, port, "granted")
			return true
		}),
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
		PublicKeyHandler: util.GeneratePublicKeyHandler(globalConfig),
	}
	log.Fatal(server.ListenAndServe())
}
