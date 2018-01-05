package daemon

import (
	"fmt"

	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/mac"
	"github.com/vishvananda/netlink"
)

// UpdateHostSideNetworkInfo will update the EndpointNetworkInfo with host side information it pulls from the host.
func UpdateHostSideNetworkInfo(in *programapi.EndpointNetworkInfo) (*programapi.EndpointNetworkInfo, error) {
	if in.IfName == "" {
		return nil, fmt.Errorf("no interface specified")
	}
	retIn := in.DeepCopy()
	if !in.ContainerMode {
		// If interface is just a host interface, IPv4 and MAC will be that of the host interface.
		link, err := netlink.LinkByName(in.IfName)
		if err != nil {
			return nil, fmt.Errorf("error while fetching Link for interface %s: %s", in.IfName, err)
		}
		if in.HostSideIfIndex != link.Attrs().Index {
			retIn.HostSideIfIndex = link.Attrs().Index
			logger.Warningf("specified host side index (%d) did not match actual value (%d). Updated!", in.HostSideIfIndex, retIn.HostSideIfIndex)
		}
		if in.HostSideMAC.String() != mac.MAC(link.Attrs().HardwareAddr).String() {
			retIn.HostSideMAC = mac.MAC(link.Attrs().HardwareAddr)
			logger.Warningf("specified host mac (%s) did not match actual value (%s). Updated!", in.HostSideMAC, retIn.HostSideMAC)
		}

		retIn.MAC = retIn.HostSideMAC
		// TODO: get IPv4 address
	} else {
		return nil, fmt.Errorf("container mode is not yet ready :(")
		// TODO: add container support for network info.
	}
	return retIn, nil
}
