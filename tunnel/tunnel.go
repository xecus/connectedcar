package tunnel

import "fmt"

type SSHServerEndpoint struct {
	Host string
	Port int
}

func (endpoint *SSHServerEndpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type PortFowardEndpoint struct {
	Host string
	Port int
}

func (endpoint *PortFowardEndpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type PortfowardSrcDst struct {
	Src *PortFowardEndpoint
	Dst *PortFowardEndpoint
}

type SSHClientConfig struct {
	User          string
	PublicKeyPath string
}

type TunnelConfig struct {
	SshServerEndpoint      *SSHServerEndpoint
	SshClientConfig        *SSHClientConfig
	LocalToRemoteForwarder []*PortfowardSrcDst
	RemoteToLocalForwarder []*PortfowardSrcDst
}

func (tc *TunnelConfig) GetLocalToRemoteForwarder() []*PortfowardSrcDst {
	return tc.LocalToRemoteForwarder
}

func (tc *TunnelConfig) GetRemoteToLocalForwarder() []*PortfowardSrcDst {
	return tc.RemoteToLocalForwarder
}
