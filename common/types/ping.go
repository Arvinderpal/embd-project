package types

import (
	"github.com/Arvinderpal/embd-project/pkg/option"
)

type PingResponse struct {
	NodeAddress string              `json:"node-address"`
	Opts        *option.BoolOptions `json:"options"`
}
