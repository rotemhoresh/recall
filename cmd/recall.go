package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rotemhoresh/recall/internal/recall"
)

// FIX: make it dynamic
const recallPath = "/home/rotemhoresh/.recall"

var (
	setFlag     string
	verboseFlag bool
)

func init() {
	flag.StringVar(&setFlag, "set", "", "set a new recall message in cwd")
	flag.BoolVar(&verboseFlag, "v", false, "verbose mode")
}

func main() {
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Print("failed to get cwd:", err)
		return
	}

	fullCWD, err := filepath.Abs(cwd)
	if err != nil {
		fmt.Print("failed to get absolute path of the cwd:", err)
		return
	}

	recalls, err := recall.Parse(recallPath)
	if err != nil {
		fmt.Print("failed to parse recalls file:", err)
		return
	}

	cwdRecallIndex := -1
	for i, recall := range recalls {
		if recall.Dir == fullCWD {
			cwdRecallIndex = i
			break
		}
	}

	if setFlag == "" { // not set
		if cwdRecallIndex == -1 {
			fmt.Println("No recall message set for this directory...")
			fmt.Println("You can set one by specifying your recall message using the -set flag.")
		} else {
			if verboseFlag {
				fmt.Println(recalls[cwdRecallIndex].Time.Format(time.ANSIC))
			}
			fmt.Println(recalls[cwdRecallIndex].Msg)
		}
	} else {
		if strings.Contains(setFlag, recall.Separator) {
			fmt.Printf("Cannot set a recall message that contains `%s`.\n", recall.Separator)
			return
		}
		if cwdRecallIndex == -1 {
			fmt.Println("No recall message set for this directory...")
			fmt.Println("Setting...")
			recalls = append(recalls, recall.Recall{
				Time: time.Now(),
				Dir:  fullCWD,
				Msg:  setFlag,
			})
		} else {
			fmt.Println("You already have a recall message for this directory, overwriting...")
			recalls[cwdRecallIndex] = recall.Recall{
				Time: time.Now(),
				Dir:  fullCWD,
				Msg:  setFlag,
			}
		}
		if err = recalls.Write(recallPath); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Done.")
	}
}
