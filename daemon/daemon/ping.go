package daemon

import (
	"github.com/Arvinderpal/embd-project/common/types"
)

func (d *Daemon) Ping() (*types.PingResponse, error) {
	d.conf.OptsMU.RLock()
	defer d.conf.OptsMU.RUnlock()
	logger.Infof("Received Ping Request...")
	return &types.PingResponse{
		NodeAddress: d.conf.NodeAddress,
		Opts:        d.conf.Opts,
	}, nil
}
