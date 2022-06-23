package snmp

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/denisbrodbeck/machineid"
	"github.com/gosnmp/gosnmp"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"ntsc.ac.cn/ta-registry/pkg/pb"
	"ntsc.ac.cn/ta-registry/pkg/rpc"
)

// TrapServer trap server
type TrapServer struct {
	conf         *TrapConfig
	trapListener *gosnmp.TrapListener
	machineID    string
	msc          pb.MonitorServiceClient
	reporter     pb.MonitorService_ReportClient
	apiServer    *echo.Echo
}

func NewTrapServer(conf *TrapConfig) (*TrapServer, error) {
	if conf == nil {
		return nil, fmt.Errorf("not set trap server config")
	}
	machineID, err := machineid.ID()
	if err != nil {
		return nil, fmt.Errorf("generate machine id failed: %v", err)
	}
	if err := conf.Check(); err != nil {
		return nil, fmt.Errorf("check config failed: %s", err.Error())
	}
	if conf.MaxOids > gosnmp.MaxOids {
		conf.MaxOids = gosnmp.MaxOids
		logrus.WithField("prefix", "trap").
			Warnf("trap max oid [%d] > %d, set to %d",
				conf.MaxOids, gosnmp.MaxOids, gosnmp.MaxOids)
	}
	uri, err := url.Parse(conf.BindAddrURI)
	if err != nil {
		return nil, fmt.Errorf("parse trap server URI failed: %s", err.Error())
	}
	port, err := strconv.Atoi(uri.Port())
	if err != nil {
		return nil, fmt.Errorf("parse URI port failed: %d", port)
	}
	if port > 65535 {
		return nil, fmt.Errorf("parse URI port failed: port must be uint16")
	}
	transType := NewTransportType(uri.Scheme)
	if transType == TRANS_UNKNOW {
		return nil, fmt.Errorf("parse URI scheme failed")
	}
	trapServer := TrapServer{
		conf:      conf,
		machineID: machineID,
	}
	trapListener := gosnmp.NewTrapListener()
	trapListener.Params = gosnmp.Default
	trapListener.Params.MaxOids = conf.MaxOids
	trapListener.Params.Community = conf.Community
	trapListener.Params.Transport = transType.String()
	trapListener.Params.Timeout = conf.Timeout
	trapListener.OnNewTrap = trapServer.TrapHandler
	trapServer.trapListener = trapListener

	tlsConf, err := rpc.GetTlsConfig(machineID, conf.CertPath, conf.ServerName)
	if err != nil {
		return nil, fmt.Errorf("generate tls config failed: %v", err)
	}
	conn, err := rpc.DialRPCConn(&rpc.DialOptions{
		RemoteAddr: conf.GatewayEndpoint,
		TLSConfig:  tlsConf,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"dial management grpc connection failed: %v", err)
	}
	trapServer.msc = pb.NewMonitorServiceClient(conn)
	e := echo.New()
	e.Debug = false
	e.HideBanner = true
	trapServer.apiServer = e
	return &trapServer, nil
}

func (s *TrapServer) Start() chan error {
	errChan := make(chan error, 1)
	c, err := s.msc.Report(context.Background())
	if err != nil {
		errChan <- fmt.Errorf("failed to create report client: %v", err)
	}
	s.reporter = c
	go s._startTrapServer(errChan)
	go s._startHttpServer(errChan)
	return errChan
}

func (s *TrapServer) _startHttpServer(errChan chan error) {
	logrus.WithField("prefix", "trap").
		Infof("start http trap server with [%s]", s.conf.HttpBindAddr)
	if err := s.apiServer.Start(s.conf.HttpBindAddr); err != nil {
		errChan <- err
		return
	}
}

func (s *TrapServer) _startTrapServer(errChan chan error) {
	logrus.WithField("prefix", "trap").
		Infof("start SNMP trap server with [%s]", s.conf.BindAddrURI)
	if err := s.trapListener.Listen(s.conf.BindAddrURI); err != nil {
		errChan <- err
		return
	}
}
