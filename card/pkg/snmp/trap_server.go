package snmp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/gosnmp/gosnmp"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"ntsc.ac.cn/ta-registry/pkg/pb"
	"ntsc.ac.cn/ta-registry/pkg/rpc"
)

// TrapServer trap server
type TrapServer struct {
	conf         *TrapConfig
	trapListener *gosnmp.TrapListener
	machineID    string
	apiServer    *echo.Echo
	grpcEntry    *grpcEntry
}

type grpcEntry struct {
	tlsConf  *tls.Config
	conn     *grpc.ClientConn
	msc      pb.MonitorServiceClient
	hc       pb.HealthClient
	reporter pb.MonitorService_ReportClient
	hwc      pb.Health_WatchClient
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
	tlsConf, err := rpc.GetTlsConfig(machineID, conf.CertPath, conf.ServerName)
	if err != nil {
		return nil, fmt.Errorf("generate tls config failed: %v", err)
	}
	trapServer := TrapServer{
		conf:      conf,
		machineID: machineID,
		grpcEntry: &grpcEntry{
			tlsConf: tlsConf,
		},
	}
	trapListener := gosnmp.NewTrapListener()
	trapListener.Params = gosnmp.Default
	trapListener.Params.MaxOids = conf.MaxOids
	trapListener.Params.Community = conf.Community
	trapListener.Params.Transport = transType.String()
	trapListener.Params.Timeout = conf.Timeout
	trapListener.OnNewTrap = trapServer.TrapHandler
	trapServer.trapListener = trapListener

	e := echo.New()
	e.Debug = false
	e.HideBanner = true
	trapServer.apiServer = e
	return &trapServer, nil
}

func (s *TrapServer) Start() chan error {
	errChan := make(chan error, 1)
	s._createRPCClient(errChan)
	go s._startHealthChekck(errChan)
	go s._startTrapServer(errChan)
	go s._startHttpServer(errChan)
	return errChan
}
func (s *TrapServer) _createRPCClient(errChan chan error) {
	var err error
	if s.grpcEntry.conn != nil {
		s.grpcEntry.conn.Close()
		s.grpcEntry.conn = nil
	}
	if s.grpcEntry.conn, err = rpc.DialRPCConn(&rpc.DialOptions{
		RemoteAddr: s.conf.GatewayEndpoint,
		TLSConfig:  s.grpcEntry.tlsConf,
	}); err != nil {
		errChan <- fmt.Errorf(
			"dial management grpc connection failed: %v", err)
		return
	}
	s.grpcEntry.msc = pb.NewMonitorServiceClient(s.grpcEntry.conn)
	s.grpcEntry.hc = pb.NewHealthClient(s.grpcEntry.conn)
	if s.grpcEntry.reporter == nil {
		if s.grpcEntry.reporter, err = s.grpcEntry.msc.Report(context.Background()); err != nil {
			logrus.WithField("prefix", "trap").
				Errorf("failed to create report client: %v", err)
			time.Sleep(time.Second)
			s._createRPCClient(errChan)
		}
	}
	if s.grpcEntry.hwc == nil {
		if s.grpcEntry.hwc, err = s.grpcEntry.hc.Watch(context.Background(),
			&pb.HealthCheckRequest{
				Service: "snmp-trap-service",
			}); err != nil {
			logrus.WithField("prefix", "trap").
				Errorf("failed to create report client: %v", err)
			time.Sleep(time.Second)
			s._createRPCClient(errChan)
		}
	}
}

func (s *TrapServer) _startHealthChekck(errChan chan error) {
	fmt.Println("start check")
	for {
		resp, err := s.grpcEntry.hwc.Recv()
		if err != nil || resp == nil {
			if strings.Contains(err.Error(), "EOF") {
				logrus.WithField("prefix", "trap").
					Errorf("snmp trap service down: %s", s.conf.HttpBindAddr)
				s.grpcEntry.reporter = nil
				s.grpcEntry.hwc = nil
				s._createRPCClient(errChan)
				continue
			}
		}
		if resp.Status != pb.HealthCheckResponse_SERVING {
			logrus.WithField("prefix", "trap").
				Warnf("snmp trap service status: %s", resp.Status)
		}
	}
}

func (s *TrapServer) _startHttpServer(errChan chan error) {
	s.apiServer.POST("/pushMonitorState", s.logHandler)
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
