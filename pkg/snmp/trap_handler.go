package snmp

import (
	"fmt"
	"io"
	"net"
	"reflect"

	"github.com/gosnmp/gosnmp"
	"github.com/sirupsen/logrus"
	"ntsc.ac.cn/ta-registry/pkg/pb"
)

func (s *TrapServer) TrapHandler(pkg *gosnmp.SnmpPacket, addr *net.UDPAddr) {
	for _, pdu := range pkg.Variables {
		data, err := NewSnmpData(pdu.Name, pdu.Type, pdu.Value)
		if err != nil {
			logrus.WithField("prefix", "trap.handler").
				Warnf("create snmp data failed: %s", err.Error())
			continue
		}
		value := fmt.Sprintf("%v", pdu.Value)
		switch pdu.Value.(type) {
		case []uint8:
			value = B2S(pdu.Value.([]uint8))
		}
		if s.grpcEntry.reporter == nil {
			continue
		}
		logrus.WithField("prefix", "trap.handler").
			Tracef("snmp [%s/%s/%v] data: %v", pdu.Name, pdu.Type,
				reflect.TypeOf(pdu.Value).String(), value)
		if err := s.grpcEntry.reporter.Send(&pb.OIDRequest{
			MachineID: s.machineID,
			Oid:       string(data.OID),
			ValueType: data.ValueType.String(),
			Value:     value,
		}); err != nil {
			if err == io.EOF {
				continue
			}
			logrus.WithField("prefix", "trap.handler").
				Errorf("failed to send snmp data: %v", err)
		}
	}
}

func B2S(bs []uint8) string {
	b := make([]byte, len(bs))
	for i, v := range bs {
		b[i] = byte(v)
	}
	return string(b)
}
