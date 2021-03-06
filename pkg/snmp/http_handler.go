package snmp

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"ntsc.ac.cn/tas/tas-commons/pkg/pb"
)

type snmpLog struct {
	ID   string `json:"id"`
	Data []*struct {
		OID   string `json:"oid"`
		State int    `json:"state"`
		Type  string `json:"type"`
	} `json:"data"`
}

type resultLog struct {
	Result string `json:"result"`
}

func (s *TrapServer) logHandler(c echo.Context) error {
	var payload snmpLog
	if err := (&echo.DefaultBinder{}).BindBody(c, &payload); err != nil {
		logrus.WithField("prefix", "trap.handler").
			Errorf("failed to bind snmp json data: %v", err)
		return c.JSON(int(codes.OK), &resultLog{Result: "fail"})
	}
	for _, v := range payload.Data {
		logrus.WithField("prefix", "trap.handler").
			Tracef("http trap [%s/%s] data: %v", v.OID, v.Type, v.State)
		machineID := s.machineID
		if payload.ID != "" {
			machineID = payload.ID
		}
		if err := s.grpcEntry.reporter.Send(&pb.OIDRequest{
			MachineID: machineID,
			Oid:       v.OID,
			ValueType: v.Type,
			Value:     fmt.Sprintf("%v", v.State),
		}); err != nil {
			logrus.WithField("prefix", "trap.handler").
				Errorf("failed to send snmp data: %v", err)
			return c.JSON(int(codes.OK), &resultLog{Result: "fail"})
		}
	}
	return c.JSON(int(codes.OK), &resultLog{Result: "success"})
}
