package commands

import (
	"fmt"

	"github.com/rotemhoresh/recall/internal/recall"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "recall",
	// TODO: add some info
	RunE: func(cmd *cobra.Command, args []string) error {
		recalls, err := recall.Parse()
		if err != nil {
			return err
		}
		if msg := recalls.Msg(); msg != "" {
			fmt.Println(msg)
		}
		return recalls.Write()
	},
}

func init() {
	rootCmd.AddCommand(
		setCmd,
		rmCmd,
		modCmd,
	)
}

func Execute() error {
	return rootCmd.Execute()
}
