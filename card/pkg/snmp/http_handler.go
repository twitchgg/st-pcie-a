package snmp

import "github.com/labstack/echo/v4"

type SNMPLog struct {
	ID   string `json:"id"`
	Data []*struct {
		OID   string `json:"oid"`
		State int    `json:"state"`
		Type  string `json:"type"`
	} `json:"data"`
}

func (s *TrapServer) logHandler(c echo.Context) error {
	var payload SNMPLog
	if err := (&echo.DefaultBinder{}).BindBody(c, &payload); err != nil {
		return nil
	}
	return nil
}
