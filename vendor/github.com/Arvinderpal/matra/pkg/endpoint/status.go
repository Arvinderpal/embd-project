package endpoint

import (
	"fmt"

	"github.com/Arvinderpal/matra/common"
)

type StatusCode int

const (
	OK       StatusCode = 0
	Warning  StatusCode = -1
	Failure  StatusCode = -2
	Disabled StatusCode = -3
	Info     StatusCode = -4
)

func NewStatusOK(info string) Status {
	return Status{Code: OK, Msg: info}
}

type Status struct {
	Code StatusCode `json:"code"`
	Msg  string     `json:"msg"`
}

func (sc StatusCode) String() string {
	var text string
	switch sc {
	case OK:
		text = common.Green("OK")
	case Warning:
		text = common.Yellow("Warning")
	case Failure:
		text = common.Red("Failure")
	case Disabled:
		text = common.Yellow("Disabled")
	case Info:
		text = common.Blue("Info")
	default:
		text = "Unknown code"
	}
	return fmt.Sprintf("%s", text)
}

func (s Status) String() string {
	if s.Msg == "" {
		return fmt.Sprintf("%s", s.Code)
	}
	return fmt.Sprintf("%s - %s", s.Code, s.Msg)
}

type StatusResponse struct {
	Logstash Status `json:"logstash"`
	Matra    Status `json:"matra"`
}
