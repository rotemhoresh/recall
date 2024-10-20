package commands

import (
	"errors"

	"github.com/rotemhoresh/recall/internal/recall"
	"github.com/spf13/cobra"
)

var modCmd = &cobra.Command{
	Use:   "mod [msg]",
	Short: "modify the recall for the cwd",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		recalls, err := recall.Parse()
		if err != nil {
			return err
		}
		if !recalls.CwdHasRecall() {
			return errors.New("cannot modify the recall for cwd as there is no recall set")
		}
		msg, err := recalls.Msg()
		if err != nil {
			return err
		}
		msg, err = fileInput(recalls.Path(), msg)
		if err != nil {
			return err
		}
		recalls.Set(msg)
		return recalls.Write()
	},
}
