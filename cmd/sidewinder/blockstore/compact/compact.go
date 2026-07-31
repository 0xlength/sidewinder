//go:build !lite

package compact

import (
	"github.com/0xlength/sidewinder/pkg/blockstore"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var Cmd = cobra.Command{
	Use:   "compact <blockstore>",
	Short: "Compact Pebble blockstore",
	Args:  cobra.ExactArgs(1),
}

func init() {
	Cmd.Run = run
}

func run(_ *cobra.Command, args []string) {
	db, err := blockstore.OpenReadWrite(args[0])
	if err != nil {
		klog.Exitf("Failed to open blockstore: %s", err)
	}
	defer db.Close()

	klog.Infof("Flushing")
	if err := db.Flush(); err != nil {
		klog.Exitf("Failed to flush: %s", err)
	}
	klog.Infof("Flushed")

	for _, cf := range db.Columns() {
		klog.Infof("Compacting %s", cf.Name)
	}
	if err := db.Compact(); err != nil {
		klog.Exitf("Failed to compact: %s", err)
	}

	klog.Infof("Done")
}
