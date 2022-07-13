package snmp

import (
	"fmt"

	"github.com/gosnmp/gosnmp"
)

const (
	OID_11D01 = ".1.3.6.1.4.1.326.1.4.1.1"
	OID_11D02 = ".1.3.6.1.4.1.326.1.4.1.2"
	OID_11D03 = ".1.3.6.1.4.1.326.1.4.1.3"
	OID_11D04 = ".1.3.6.1.4.1.326.1.4.1.4"
	OID_11D05 = ".1.3.6.1.4.1.326.1.4.1.5"
	OID_11D06 = ".1.3.6.1.4.1.326.1.4.1.6"
	OID_11D07 = ".1.3.6.1.4.1.326.1.4.1.7"
	OID_11D08 = ".1.3.6.1.4.1.326.1.4.1.8"
	OID_11D09 = ".1.3.6.1.4.1.326.1.4.1.9"
	OID_11D10 = ".1.3.6.1.4.1.326.1.4.1.10"
	OID_11D11 = ".1.3.6.1.4.1.326.1.4.1.11"
	OID_11D12 = ".1.3.6.1.4.1.326.1.4.1.12"
	OID_11D13 = ".1.3.6.1.4.1.326.1.4.1.13"
	OID_11D14 = ".1.3.6.1.4.1.326.1.4.1.14"
	OID_11D15 = ".1.3.6.1.4.1.326.1.4.1.15"
	OID_11D16 = ".1.3.6.1.4.1.326.1.4.1.16"
	OID_11D17 = ".1.3.6.1.4.1.326.1.4.1.17"
	OID_11D18 = ".1.3.6.1.4.1.326.1.4.1.18"
	OID_11D19 = ".1.3.6.1.4.1.326.1.4.1.19"
	OID_11D20 = ".1.3.6.1.4.1.326.1.4.1.20"
)

type SnmpData struct {
	OID       OID_11D
	ValueType gosnmp.Asn1BER
	Value     interface{}
}

func NewSnmpData(oid string, vt gosnmp.Asn1BER, v interface{}) (*SnmpData, error) {
	aoid := OID_11D(oid)
	return &SnmpData{
		OID:       aoid,
		ValueType: vt,
		Value:     v,
	}, nil
}

type OID_TYPE int

const (
	OID_STATUS OID_TYPE = iota
	OID_SETUP
)

func (ot OID_TYPE) String() string {
	switch ot {
	case OID_STATUS:
		return "status"
	case OID_SETUP:
		return "setup"
	default:
		return ""
	}
}

type OID_11D string

func (oid OID_11D) Type() OID_TYPE {
	switch oid {
	case OID_11D07, OID_11D08, OID_11D09, OID_11D10:
		return OID_SETUP
	case OID_11D16, OID_11D17, OID_11D18, OID_11D19, OID_11D20:
		return OID_SETUP
	default:
		return OID_STATUS
	}
}

func (oid OID_11D) Desc() string {
	switch oid {
	case OID_11D01:
		return "监控方式"
	case OID_11D02:
		return "同步精度"
	case OID_11D03:
		return "同步指示"
	case OID_11D04:
		return "工作状态"
	case OID_11D05:
		return "输入状态"
	case OID_11D06:
		return "参考源状态"
	case OID_11D07:
		return "网管-IP地址"
	case OID_11D08:
		return "网管-子网"
	case OID_11D09:
		return "网管-网关"
	case OID_11D10:
		return "中断使能"
	case OID_11D11:
		return "设备描述"
	case OID_11D12:
		return "设备标识"
	case OID_11D13:
		return "设备运行时间"
	case OID_11D14:
		return "设备守时时间"
	case OID_11D15:
		return "联系人"
	case OID_11D16:
		return "标准时间"
	case OID_11D17:
		return "时区选择"
	case OID_11D18:
		return "闰秒预告"
	case OID_11D19:
		return "时延补偿"
	case OID_11D20:
		return "保留"
	default:
		return string(oid)
	}
}

func (sd *SnmpData) String() string {
	switch sd.ValueType {
	case gosnmp.OctetString:
		return fmt.Sprintf("oid [%s][%s],value: %v", sd.OID.Desc(), sd.OID, string(sd.Value.([]byte)))
	case gosnmp.Integer:
		return fmt.Sprintf("oid [%s][%s],value: %d", sd.OID.Desc(), sd.OID, sd.Value.(int))
	default:
		return fmt.Sprintf("oid [%s][%s],type [%s], value: %v", sd.OID.Desc(), sd.OID, sd.ValueType.String(), sd.Value)
	}
}
