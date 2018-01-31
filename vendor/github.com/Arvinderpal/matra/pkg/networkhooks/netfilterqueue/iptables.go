package netfilterqueue

import (
	"os/exec"
	"strconv"
	"strings"
)

const iptablesCommand = "iptables"

// addIptablesRules add iptable rules based on user config so that packets
// are routed from the network interface to the desired queue.

func (h *NetfilterQueue) addIptablesRules() error {

	// Delete any existing rules with the same signature before adding new rules
	h.removeIptablesRules()

	// TODO: add FORWARD support
	if h.State.Conf.NFQConf.PacketFilter != nil && strings.ToUpper(h.State.Conf.NFQConf.PacketFilter.IptablesChain) == "INPUT" {
		// Example:
		// iptables -A INPUT -i eth0 -p icmp -j NFQUEUE --queue-num 200
		args := buildIptablesInputArgs(true, h.State.Conf.NFQConf)
		cmd := exec.Command(iptablesCommand, args...)
		err := debugCommandOutput(cmd)
		if err != nil {
			return err
		}

	} else if h.State.Conf.NFQConf.PacketFilter != nil && strings.ToUpper(h.State.Conf.NFQConf.PacketFilter.IptablesChain) == "OUTPUT" {
		// Example:
		// iptables -A OUTPUT -o eth0 -p icmp -j NFQUEUE --queue-num 200
		args := buildIptablesOutputArgs(true, h.State.Conf.NFQConf)
		cmd := exec.Command(iptablesCommand, args...)
		err := debugCommandOutput(cmd)
		if err != nil {
			return err
		}
	} else { // (default) both INPUT and OUTPUT
		args := buildIptablesInputArgs(true, h.State.Conf.NFQConf)
		cmd := exec.Command(iptablesCommand, args...)
		err := debugCommandOutput(cmd)
		if err != nil {
			return err
		}

		args = buildIptablesOutputArgs(true, h.State.Conf.NFQConf)
		cmd = exec.Command(iptablesCommand, args...)
		err = debugCommandOutput(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *NetfilterQueue) removeIptablesRules() error {
	// TODO: add FORWARD support
	if h.State.Conf.NFQConf.PacketFilter != nil && strings.ToUpper(h.State.Conf.NFQConf.PacketFilter.IptablesChain) == "INPUT" {
		// Example:
		// iptables -A INPUT -i eth0 -p icmp -j NFQUEUE --queue-num 200
		args := buildIptablesInputArgs(false, h.State.Conf.NFQConf)
		cmd := exec.Command(iptablesCommand, args...)
		err := debugCommandOutput(cmd)
		if err != nil {
			return err
		}

	} else if h.State.Conf.NFQConf.PacketFilter != nil && strings.ToUpper(h.State.Conf.NFQConf.PacketFilter.IptablesChain) == "OUTPUT" {
		// Example:
		// iptables -A OUTPUT -o eth0 -p icmp -j NFQUEUE --queue-num 200
		args := buildIptablesOutputArgs(false, h.State.Conf.NFQConf)
		cmd := exec.Command(iptablesCommand, args...)
		err := debugCommandOutput(cmd)
		if err != nil {
			return err
		}
	} else { // (default) both INPUT and OUTPUT
		args := buildIptablesInputArgs(false, h.State.Conf.NFQConf)
		cmd := exec.Command(iptablesCommand, args...)
		err := debugCommandOutput(cmd)
		if err != nil {
			return err
		}

		args = buildIptablesOutputArgs(false, h.State.Conf.NFQConf)
		cmd = exec.Command(iptablesCommand, args...)
		err = debugCommandOutput(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildIptablesInputArgs(addRules bool, nfqconf NFQueueConf) []string {
	var args []string
	var tmp []string
	if addRules {
		args = append(args, "-A")
	} else {
		args = append(args, "-D")
	}
	args = append(args, "INPUT")
	args = append(args, "-i")
	args = append(args, nfqconf.Iface)

	tmp = buildIptablesProtocolPortArgs(nfqconf.PacketFilter)
	args = append(args, tmp...)
	tmp = buildIptablesNFQueueArgs(nfqconf)
	args = append(args, tmp...)

	return args
}
func buildIptablesOutputArgs(addRules bool, nfqconf NFQueueConf) []string {
	var args []string
	var tmp []string

	if addRules {
		args = append(args, "-A")
	} else {
		args = append(args, "-D")
	}
	args = append(args, "OUTPUT")
	args = append(args, "-o")
	args = append(args, nfqconf.Iface)

	tmp = buildIptablesProtocolPortArgs(nfqconf.PacketFilter)
	args = append(args, tmp...)
	tmp = buildIptablesNFQueueArgs(nfqconf)
	args = append(args, tmp...)
	return args
}

func buildIptablesNFQueueArgs(nfqconf NFQueueConf) []string {
	var args []string
	args = append(args, "-j")
	args = append(args, "NFQUEUE")
	args = append(args, "--queue-num")
	args = append(args, strconv.Itoa(nfqconf.QueueId))
	return args
}

func buildIptablesProtocolPortArgs(filter *PacketFilterConf) []string {
	var args []string
	if filter != nil && filter.ProtocolPort != nil {
		if filter.ProtocolPort.Protocol != "" {
			args = append(args, "-p")
			args = append(args, filter.ProtocolPort.Protocol)
		}
		if filter.ProtocolPort.Port != nil {
			if filter.ProtocolPort.Source {
				// iptables -A INPUT -i eth0 -p tcp --sport 22
				args = append(args, "--sport")
			} else {
				args = append(args, "--dport")
			}
			args = append(args, strconv.Itoa(*filter.ProtocolPort.Port))
		}
	}
	return args
}

// Utility to debug the output from commands executed by netfilter.
func debugCommandOutput(cmd *exec.Cmd) (err error) {
	logger.Infof("Executing: %s %s", cmd.Path, strings.Join(cmd.Args[1:], " "))
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("%s error: %s %q", cmd.Path, err, string(out))
	}
	return err
}
