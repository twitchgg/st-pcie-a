package statistics

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type HWstatistics struct {
}

func NewHWStatistics() *HWstatistics {
	return &HWstatistics{}
}

func (*HWstatistics) CPUPercent() (int, error) {
	ps, err := cpu.Percent(time.Second, true)
	if err != nil {
		return 0, err
	}
	return int(ps[0] * 100), nil
}

func (*HWstatistics) MemUsedPercent() (int, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return 0, nil
	}
	return int(stat.UsedPercent * 100), nil
}

func (*HWstatistics) DiskRootUsedPercent() (int, error) {
	stat, err := disk.Usage("/")
	if err != nil {
		return 0, nil
	}
	return int(stat.UsedPercent * 100), nil
}

func (*HWstatistics) NetIOCounters() (sent int, recv int, err error) {
	var n1, n2 []net.IOCountersStat
	if n1, err = net.IOCounters(false); err != nil {
		return
	}
	time.Sleep(time.Second)
	if n2, err = net.IOCounters(false); err != nil {
		return
	}
	sent = int(n2[0].BytesSent - n1[0].BytesSent)
	recv = int(n2[0].BytesRecv - n1[0].BytesRecv)
	return
}
