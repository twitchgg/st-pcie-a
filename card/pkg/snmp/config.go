package snmp

import (
	"strings"
	"time"
)

// TransportType transport type (tcp/udp)
type TransportType int

const (
	TRANS_UNKNOW TransportType = iota
	TRANS_UDP
	TRANS_TCP
)

const (
	CERT_EXT_KEY_MACHINE_ID = "1.1.1.1.1.1"
	TRUSTED_CERT_CHAIN_NAME = "trusted.crt"
	CLIENT_CERT_NAME        = "client.crt"
	CLIENT_PRIVATE_KEY_NAME = "client.key"
)

func (t TransportType) String() string {
	switch t {
	case TRANS_UDP:
		return "udp"
	case TRANS_TCP:
		return "tcp"
	default:
		return ""
	}
}

func NewTransportType(t string) TransportType {
	switch strings.ToLower(t) {
	case "udp":
		return TRANS_UDP
	case "tcp":
		return TRANS_TCP
	default:
		return TRANS_UNKNOW
	}
}

// TrapConfig trap server config
type TrapConfig struct {
	// BindAddr binding address uri (default udp://127.0.0.1:1169)
	BindAddrURI        string `default:"udp://127.0.0.1:1169"`
	Community          string `default:"1234qwer"`
	ExponentialTimeout bool
	MaxOids            int
	Timeout            time.Duration
	CertPath           string
	ServerName         string
	GatewayEndpoint    string
}

// Check check config
func (tc *TrapConfig) Check() error {
	return nil
}
