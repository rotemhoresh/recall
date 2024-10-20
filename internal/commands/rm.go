package commands

import (
	"github.com/rotemhoresh/recall/internal/recall"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "makes sure there is no recall for cwd.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		recalls, err := recall.Parse()
		if err != nil {
			return err
		}
		recalls.Delete()
		return recalls.Write()
	},
}
