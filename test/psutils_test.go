package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

func TestCPU(t *testing.T) {
	fmt.Println(cpu.Percent(time.Second, false))
	stat, err := mem.VirtualMemory()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(stat.UsedPercent, int(stat.UsedPercent*100))
	stat1, err := disk.Usage("/")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(stat1.Total, stat1.Used, stat1.UsedPercent)
	n1, err := net.IOCounters(false)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	n2, err := net.IOCounters(false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("BytesSent:", n2[0].BytesSent-n1[0].BytesSent,
		"BytesRecv:", n2[0].BytesRecv-n1[0].BytesRecv)
}
