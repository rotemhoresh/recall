package commands

import (
	"os"
	"os/exec"

	"github.com/rotemhoresh/recall/internal/recall"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [msg]",
	Short: "sets a recall for the cwd",
	Long: `sets a recall for the current working directory. 
  if there is already a recall set for this cwd, it will be overwritten.
  if no msg is specified, an editor will be opened to set the message in.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		recalls, err := recall.Parse()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			msg, err := fileInput(recalls.Path(), "")
			if err != nil {
				return err
			}
			recalls.Set(msg)
		} else {
			recalls.Set(args[0])
		}
		return recalls.Write()
	},
}

func fileInput(path, initial string) (string, error) {
	// TODO: add option for a config file to allow editor selection etc.
	filePath := path + "/EDITMSG"

	err := os.WriteFile(filePath, []byte(initial), 0o644)
	if err != nil {
		return "", err
	}
	cmd := exec.Command("nvim", filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	f, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(f), nil
}
