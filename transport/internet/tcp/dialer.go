// +build !confonly

package tcp

import (
	"context"

	"github.com/xtls/xray-core/v1/common"
	"github.com/xtls/xray-core/v1/common/net"
	"github.com/xtls/xray-core/v1/common/session"
	"github.com/xtls/xray-core/v1/transport/internet"
	"github.com/xtls/xray-core/v1/transport/internet/tls"
	"github.com/xtls/xray-core/v1/transport/internet/xtls"
)

// Dial dials a new TCP connection to the given destination.
func Dial(ctx context.Context, dest net.Destination, streamSettings *internet.MemoryStreamConfig) (internet.Connection, error) {
	newError("dialing TCP to ", dest).WriteToLog(session.ExportIDToError(ctx))
	conn, err := internet.DialSystem(ctx, dest, streamSettings.SocketSettings)
	if err != nil {
		return nil, err
	}

	if config := tls.ConfigFromStreamSettings(streamSettings); config != nil {
		tlsConfig := config.GetTLSConfig(tls.WithDestination(dest))
		/*
			if config.IsExperiment8357() {
				conn = tls.UClient(conn, tlsConfig)
			} else {
				conn = tls.Client(conn, tlsConfig)
			}
		*/
		conn = tls.Client(conn, tlsConfig)
	} else if config := xtls.ConfigFromStreamSettings(streamSettings); config != nil {
		xtlsConfig := config.GetXTLSConfig(xtls.WithDestination(dest))
		conn = xtls.Client(conn, xtlsConfig)
	}

	tcpSettings := streamSettings.ProtocolSettings.(*Config)
	if tcpSettings.HeaderSettings != nil {
		headerConfig, err := tcpSettings.HeaderSettings.GetInstance()
		if err != nil {
			return nil, newError("failed to get header settings").Base(err).AtError()
		}
		auth, err := internet.CreateConnectionAuthenticator(headerConfig)
		if err != nil {
			return nil, newError("failed to create header authenticator").Base(err).AtError()
		}
		conn = auth.Client(conn)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(protocolName, Dial))
}