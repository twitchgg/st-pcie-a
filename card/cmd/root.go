package cmd

import (
	"flag"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"ntsc.ac.cn/st-pcie-a/card/pkg/snmp"
)

var envs struct {
	bindAddr      string
	community     string
	ntpServerAddr string
}

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	formatter := new(prefixed.TextFormatter)
	logrus.SetFormatter(formatter)
	flag.StringVar(&envs.bindAddr, "snmp-bind-addr",
		"udp://127.0.0.1:1169", "SNMP binding addresss")
	flag.StringVar(&envs.community, "snmp-community",
		"1234qwer", "SNMP community string")
	flag.StringVar(&envs.ntpServerAddr, "ntp-addr",
		"udp://ntp1.aliyun.com:123", "NTP server address")
}

func Execute() {
	serv, err := snmp.NewTrapServer(&snmp.TrapConfig{
		BindAddrURI:        envs.bindAddr,
		Community:          envs.community,
		ExponentialTimeout: true,
		Timeout:            time.Duration(time.Second * 3),
	})
	if err != nil {
		logrus.WithField("prefix", "main").Fatal(err.Error())
	}
	logrus.WithField("prefix", "main").Fatal(<-serv.Start())
}
