package snmp

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
)

// TrapServer trap server
type TrapServer struct {
	conf         *TrapConfig
	trapListener *gosnmp.TrapListener
}

func NewTrapServer(conf *TrapConfig) (*TrapServer, error) {
	if conf == nil {
		return nil, fmt.Errorf("not set trap server config")
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
		conf: conf,
	}
	trapListener := gosnmp.NewTrapListener()
	trapListener.Params = gosnmp.Default
	trapListener.Params.MaxOids = conf.MaxOids
	trapListener.Params.Community = conf.Community
	trapListener.Params.Transport = transType.String()
	trapListener.Params.Timeout = conf.Timeout
	trapListener.OnNewTrap = trapServer.TrapHandler
	trapServer.trapListener = trapListener
	return &trapServer, nil
}

func (s *TrapServer) Start() chan error {
	errChan := make(chan error, 1)
	go s._startTrapServer(errChan)
	return errChan
}

func (s *TrapServer) _startTrapServer(errChan chan error) {
	logrus.WithField("prefix", "trap").
		Infof("start SNMP trap server with [%s]", s.conf.BindAddrURI)
	if err := s.trapListener.Listen(s.conf.BindAddrURI); err != nil {
		errChan <- err
		return
	}
}
