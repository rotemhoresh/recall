package commands

import (
	"errors"
	"os"
	"os/exec"
	"strings"

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
			msg, err := fileInput(recalls.Path())
			if err != nil {
				return err
			}
			msg = strings.Trim(msg, " \n")
			if strings.Contains(msg, "\n") {
				return errors.New("newline characters are not allowed inside a recall")
			}
			recalls.Set(msg)
		} else {
			recalls.Set(args[0])
		}
		return recalls.Write()
	},
}

func fileInput(path string) (string, error) {
	// TODO: add option for a config file to allow editor selection etc.
	cmd := exec.Command("nvim", path+"/EDITMSG")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	f, err := os.ReadFile(path + "/EDITMSG")
	if err != nil {
		return "", err
	}
	return string(f), nil
}
