package util

//
// 主にTunnelAgentのClient側から使うパーツ群
//

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xecus/connectedcar/tunnel"
	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	tunnelConfig    *tunnel.TunnelConfig
	host            string
	config          *ssh.ClientConfig
	remoteListeners map[string]net.Listener
	mux             sync.RWMutex
	c               *ssh.Client
}

func NewSSHClient(tunnelConfig *tunnel.TunnelConfig) (*SSHClient, error) {

	host := tunnelConfig.SshServerEndpoint.String()
	config := &ssh.ClientConfig{
		// SSH connection username
		User: tunnelConfig.SshClientConfig.User,
		Auth: []ssh.AuthMethod{
			// put here your private key path
			PublicKeyFile(tunnelConfig.SshClientConfig.PublicKeyPath),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	log.Printf("Dialing to %s\n", host)
	c, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}
	log.Printf("OK")

	return &SSHClient{
		tunnelConfig:    tunnelConfig,
		host:            host,
		config:          config,
		remoteListeners: map[string]net.Listener{},
		c:               c,
	}, nil
}

func (c *SSHClient) Dial(ctx context.Context, n, addr string) (net.Conn, error) {
	conn, err := c.getC().Dial(n, addr)
	if err != nil {
		if rErr := c.reconnect(ctx); rErr != nil {
			return nil, err
		}
		return c.getC().Dial(n, addr)
	}
	return conn, nil
}

func (c *SSHClient) ListenPortOnRemote() error {
	for _, forwardInfo := range c.tunnelConfig.RemoteToLocalForwarder {
		srcHostPort := forwardInfo.Src.String()
		dstHostPort := forwardInfo.Dst.String()
		log.Printf("[ListenPort on Remote] %s -> %s\n", srcHostPort, dstHostPort)
		listener, err := c.getC().Listen("tcp", srcHostPort)
		if err != nil {
			return err
		}
		c.remoteListeners[srcHostPort] = listener
	}
	return nil
}

type acceptRoutineController struct {
	v   map[string]bool
	mux sync.Mutex
}

func (c *acceptRoutineController) Init() {
	c.v = map[string]bool{}
}
func (c *acceptRoutineController) Working(key string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.v[key] = true
}
func (c *acceptRoutineController) Stopped(key string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.v[key] = false
}

func (c *acceptRoutineController) IsWorking(key string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	if tmpStat, exists := c.v[key]; exists {
		return tmpStat
	}
	return false
}

func (c *SSHClient) BridgeLocalAndRemoteConnection(ctx context.Context) error {
	// Local -^> Remote
	go func() {

	}()

	// Remote -> Local

	go func() {
		arc := &acceptRoutineController{}
		arc.Init()

		for {
			for _, forwardInfo := range c.tunnelConfig.RemoteToLocalForwarder {

				srcHostPort := forwardInfo.Src.String()
				dstHostPort := forwardInfo.Dst.String()
				forwarderName := fmt.Sprintf("%s -> %s", srcHostPort, dstHostPort)

				// Need not to accept client for closed remote listener
				remoteListener, exists := c.remoteListeners[srcHostPort]
				if !exists {
					continue
				}

				// Need run goroutine?
				if arc.IsWorking(forwarderName) {
					continue
				}
				arc.Working(forwarderName)

				go func() {
					defer arc.Stopped(forwarderName)

					connCtx, cancel := context.WithCancel(ctx)
					defer cancel()

					// Accept
					log.Println(forwarderName, "*** Waiting Accept... ***", srcHostPort)
					remoteClient, err := remoteListener.Accept()
					if err != nil {
						log.Println(err)
						return
					}
					go func() {
						<-connCtx.Done()
						log.Println(forwarderName, "Closing remote Client...")
						remoteClient.Close()
					}()

					log.Println(forwarderName, "Accepted Client on Remote.")

					// Dial
					log.Println(forwarderName, "Dialing local port...")
					localConn, err := net.Dial("tcp", dstHostPort)
					if err != nil {
						log.Println(forwarderName, "Failed to dial local port")
						remoteClient.Close()
						return
					}
					go func() {
						<-connCtx.Done()
						log.Println(forwarderName, "Closing local connection...")
						localConn.Close()
					}()
					log.Println(forwarderName, "Connection Established.")

					// Bridge
					MakeBridgeConnection(remoteClient, localConn, cancel)
					log.Println(forwarderName, "Bridge Closed.")
				}()

			}
		}
	}()

	return nil
}

func (c *SSHClient) GetRemoteListeners() []net.Listener {
	ret := []net.Listener{}
	for _, v := range c.remoteListeners {
		ret = append(ret, v)
	}
	return ret
}

func (c *SSHClient) Close() error {
	return c.getC().Close()
}

func (c *SSHClient) KeepAlive(ctx context.Context) {
	wait := make(chan error, 1)
	go func() {
		wait <- c.getC().Wait()
	}()

	var aliveErrCount uint32
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wait:
			return
		case <-ticker.C:
			if atomic.LoadUint32(&aliveErrCount) > 1 {
				log.Printf("failed to keep alive of %v", c.getC().RemoteAddr())
				c.getC().Close()
				return
			}
		case <-ctx.Done():
			return
		}

		go func() {
			_, _, err := c.getC().SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				atomic.AddUint32(&aliveErrCount, 1)
			}
		}()
	}
}

func (c *SSHClient) getC() *ssh.Client {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.c
}

func (c *SSHClient) setC(client *ssh.Client) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.c = client
}

func (c *SSHClient) reconnect(ctx context.Context) error {
	client, err := ssh.Dial("tcp", c.host, c.config)
	if err != nil {
		return err
	}
	c.getC().Close()
	c.setC(client)
	go c.KeepAlive(ctx)
	return nil
}
