package sshclient

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/xecus/connectedcar/config"
	"golang.org/x/crypto/ssh"
)

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

// From https://sosedoff.com/2015/05/25/ssh-port-forwarding-with-go.html
// Handle local client connections and tunnel data to the remote server
// Will use io.Copy - http://golang.org/pkg/io/#Copy
func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
		}
		chDone <- true
	}()

	<-chDone
}

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot read SSH public key file %s", file))
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot parse SSH public key file %s", file))
		return nil
	}
	return ssh.PublicKeys(key)
}

// local service to be forwarded
var localEndpoint = Endpoint{
	Host: "localhost",
	Port: 8090,
}

// remote SSH server
var serverEndpoint = Endpoint{
	//Host: "34.146.73.112",
	Host: "localhost",
	Port: 2222,
}

// remote forwarding port (on remote SSH server network)
var remoteEndpoint = Endpoint{
	Host: "localhost",
	Port: 8080,
}

func SshConnectionClient(globalConfig *config.GlobalConfig) {

	// refer to https://godoc.org/golang.org/x/crypto/ssh for other authentication types
	sshConfig := &ssh.ClientConfig{
		// SSH connection username
		User: "admin+xxx",
		Auth: []ssh.AuthMethod{
			// put here your private key path
			publicKeyFile("/home/hiroyuki/id_rsa"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to SSH remote server using serverEndpoint
	log.Println("Dialing...", serverEndpoint.String())
	serverConn, err := ssh.Dial("tcp", serverEndpoint.String(), sshConfig)
	if err != nil {
		log.Fatalln(fmt.Printf("Dial INTO remote server error: %s", err))
	}
	log.Println("OK")

	// Listen on remote server port
	log.Println("Listen on remote server port...")
	listener, err := serverConn.Listen("tcp", remoteEndpoint.String())
	if err != nil {
		log.Fatalln(fmt.Printf("Listen open port ON remote server error: %s", err))
	}
	defer listener.Close()
	log.Println("OK")

	// handle incoming connections on reverse forwarded tunnel
	for {
		// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
		log.Println("Open Local Connection...", localEndpoint.String())
		local, err := net.Dial("tcp", localEndpoint.String())
		if err != nil {
			//log.Fatalln(fmt.Printf("Dial INTO local service error: %s", err))
			log.Println(fmt.Printf("Dial INTO local service error: %s", err))
			time.Sleep(1.0)
			continue
		}
		log.Println("OK")

		log.Println("Open Listener...")
		client, err := listener.Accept()
		if err != nil {
			//log.Fatalln(err)
			log.Println(err)
			time.Sleep(1.0)
			continue
		}
		log.Println("OK")

		handleClient(client, local)
	}

}
