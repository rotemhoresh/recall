package commands

import (
	"github.com/rotemhoresh/recall/internal/recall"
	"github.com/spf13/cobra"
)

var modCmd = &cobra.Command{
	Use:   "mod [msg]",
	Short: "modify the recall for the cwd",
	Long: `modify the recall for the current working directory.
if none is set, a new recall will be created.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		recalls, err := recall.Parse()
		if err != nil {
			return err
		}
		msg, err := fileInput(recalls.Path(), recalls.Msg())
		if err != nil {
			return err
		}
		recalls.Set(msg)
		return recalls.Write()
	},
}
