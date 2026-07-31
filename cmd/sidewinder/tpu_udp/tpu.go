package tpu_udp

import (
	"github.com/0xlength/sidewinder/cmd/sidewinder/tpu_udp/pcap"
	"github.com/0xlength/sidewinder/cmd/sidewinder/tpu_udp/proxy"
	"github.com/0xlength/sidewinder/cmd/sidewinder/tpu_udp/sniff"
	"github.com/spf13/cobra"
)

var Cmd = cobra.Command{
	Use:   "tpu-udp",
	Short: "TPU/UDP tools",
}

func init() {
	Cmd.AddCommand(
		&pcap.Cmd,
		&proxy.Cmd,
		&sniff.Cmd,
	)
}
