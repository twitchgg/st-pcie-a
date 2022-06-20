package snmp

import (
	"net"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
)

func (s *TrapServer) TrapHandler(pkg *gosnmp.SnmpPacket, addr *net.UDPAddr) {
	// logrus.WithField("prefix", "trap.Handler").
	// Debugf("received from: %s,pdu size [%d]", addr.IP.String(), len(pkg.Variables))
	for _, pdu := range pkg.Variables {
		data, err := NewSnmpData(pdu.Name, pdu.Type, pdu.Value)
		if err != nil {
			logrus.WithField("prefix", "trap.handler").
				Warnf("create snmp data failed: %s", err.Error())
			continue
		}
		logrus.WithField("prefix", "trap.handler").Trace(data.String())
	}
}
