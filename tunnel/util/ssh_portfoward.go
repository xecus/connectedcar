package util

import (
	"io"
	"net"
)

// From https://sosedoff.com/2015/05/25/ssh-port-forwarding-with-go.html
// Handle local client connections and tunnel data to the remote server
// Will use io.Copy - http://golang.org/pkg/io/#Copy

func MakeBridgeConnection(client net.Conn, remote net.Conn, shutdown func()) {
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			//log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
			shutdown()
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			//log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
			shutdown()
		}
		chDone <- true
	}()
	<-chDone
}
