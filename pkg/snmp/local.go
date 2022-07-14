package snmp

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"ntsc.ac.cn/tas/tas-commons/pkg/pb"
)

func (s *TrapServer) localTrapJobFunc() {
	cpu, err := s.hws.CPUPercent()
	if err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to read cpu percent: %v", err)
		return
	}
	if err = s.sendTrap(".1.3.6.1.4.1.326.2.1.1.1",
		"Integer", cpu); err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to send cpu percent: %v", err)
		return
	}
	mem, err := s.hws.MemUsedPercent()
	if err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to read memory used percent: %v", err)
		return
	}
	if err = s.sendTrap(".1.3.6.1.4.1.326.2.1.1.2",
		"Integer", mem); err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to send memory used percent: %v", err)
		return
	}
	disk, err := s.hws.DiskRootUsedPercent()
	if err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to read disk root used percent: %v", err)
		return
	}
	if err = s.sendTrap(".1.3.6.1.4.1.326.2.1.1.3",
		"Integer", disk); err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to send disk root used percent: %v", err)
		return
	}
	sent, recv, err := s.hws.NetIOCounters()
	if err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to read net io counters: %v", err)
		return
	}
	if err = s.sendTrap(".1.3.6.1.4.1.326.2.1.1.4",
		"Counter64", sent); err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to send net sent bytes: %v", err)
		return
	}
	if err = s.sendTrap(".1.3.6.1.4.1.326.2.1.1.5",
		"Counter64", recv); err != nil {
		logrus.WithField("prefix", "trap.local").
			Errorf("failed to send net recv bytes: %v", err)
		return
	}
}
func (s *TrapServer) _startLocalTrap(errChan chan error) {
	s.crontab.AddFunc("@every 10s", s.localTrapJobFunc)
	s.crontab.Start()
}

func (s *TrapServer) sendTrap(oid, vt string, v int) error {
	if err := s.grpcEntry.reporter.Send(&pb.OIDRequest{
		MachineID: s.machineID,
		Oid:       oid,
		ValueType: vt,
		Value:     fmt.Sprintf("%v", v),
	}); err != nil {
		logrus.WithField("prefix", "trap.handler").
			Errorf("failed to send snmp data: %v", err)
		return err
	}
	logrus.WithField("prefix", "trap.local").
		Tracef("send local trap oid [%s] type [%s] value [%v]", oid, vt, v)
	return nil
}
