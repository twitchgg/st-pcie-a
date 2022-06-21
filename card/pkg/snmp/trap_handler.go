package snmp

import (
	"fmt"
	"net"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"ntsc.ac.cn/ta-registry/pkg/pb"
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
		if err := s.reporter.Send(&pb.OIDRequest{
			MachineID: s.machineID,
			Oid:       string(data.OID),
			ValueType: data.ValueType.String(),
			Value:     fmt.Sprintf("%v", data.Value),
		}); err != nil {
			logrus.WithField("prefix", "trap.handler").
				Errorf("failed to send snmp data: %v", err)
		}
	}
}
