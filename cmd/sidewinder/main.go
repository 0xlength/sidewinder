package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/0xlength/sidewinder/cmd/sidewinder/tpu_quic"
	"github.com/0xlength/sidewinder/cmd/sidewinder/tpu_udp"

	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore"
	"github.com/0xlength/sidewinder/cmd/sidewinder/gossip"
	"github.com/0xlength/sidewinder/cmd/sidewinder/replay"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	// Load in instruction pretty-printing
	_ "github.com/gagliardetto/solana-go/programs/system"
	_ "github.com/gagliardetto/solana-go/programs/vote"
)

var cmd = cobra.Command{
	Use:   "sidewinder",
	Short: "Solana Go playground",
}

func init() {
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	cmd.PersistentFlags().AddGoFlagSet(klogFlags)

	cmd.AddCommand(
		&blockstore.Cmd,
		&gossip.Cmd,
		&replay.Cmd,
		&tpu_udp.Cmd,
		&tpu_quic.Cmd,
	)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	cobra.CheckErr(cmd.ExecuteContext(ctx))
}
