// package recall handles storing and parsing recalls.
//
// The format of the recalls file is like the following line, with each like
// accounting for a [Recall], recall segments are separated with a ` ! ` and lines
// are separated with a `\n`.
//
//	<time, in [time.RFC3339] format> ! <dir path> ! <message>
package recall

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rotemhoresh/recall/internal/pathutils"
)

const Separator = " ! "

type Recall struct {
	Time time.Time
	Dir  string
	Msg  string
}

type Recalls struct {
	cwdIndex int // the index of the recall of the cwd
	cwd      string
	recalls  []Recall
	path     string // $(HOME)/.recall
}

type ParsingError string

func (e ParsingError) Error() string {
	return fmt.Sprintf("recall: parsing error: %s", string(e))
}

// Parse parses a file of recalls.
func Parse() (Recalls, error) {
	cwd, err := pathutils.Cwd()
	if err != nil {
		return Recalls{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return Recalls{}, err
	}
	path := fmt.Sprintf("%s/.recall", home)
	recalls, cwdIndex, err := parseRecalls(path+"/recalls", cwd)
	if err != nil {
		return Recalls{}, err
	}
	return Recalls{
		recalls:  recalls,
		cwdIndex: cwdIndex,
		path:     path,
		cwd:      cwd,
	}, nil
}

func parseRecalls(path string, cwd string) ([]Recall, int, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	lines := strings.Split(string(f), "\n")
	lines = lines[:len(lines)-1] // last one is always empty
	// if the file is in a valid format, each line accounts for a recall
	recalls := make([]Recall, 0, len(lines))

	cwdIndex := -1

	for i, ln := range lines {
		segments := strings.Split(ln, Separator)
		if l := len(segments); l != 3 {
			return nil, 0, ParsingError(fmt.Sprintf("expected line to have %d segments, but got %d", 3, l))
		}
		t, err := time.Parse(time.RFC3339, segments[0])
		if err != nil {
			return nil, 0, ParsingError(err.Error())
		}
		if !pathutils.Valid(segments[1]) {
			return nil, 0, ParsingError(fmt.Sprintf("invalid path format: %s", segments[1]))
		}
		if dirExists, err := pathutils.DirExists(segments[1]); err != nil {
			return nil, 0, err
		} else if !dirExists {
			// if the directory doesn't exist anymore, the recall message is no longer
			// needed, thus we are not adding this recall to the recalls, and it will be
			// deleted when the new recalls we be written to the file.
			continue
		}
		recalls = append(recalls, Recall{
			Time: t,
			Dir:  segments[1],
			Msg:  segments[2],
		})
		if segments[1] == cwd {
			cwdIndex = i
		}
	}

	return recalls, cwdIndex, nil
}

func (r *Recalls) Write() error {
	var b bytes.Buffer
	for _, recall := range r.recalls {
		if _, err := b.WriteString(fmt.Sprintf("%s ! %s ! %s\n", recall.Time.Format(time.RFC3339), recall.Dir, recall.Msg)); err != nil {
			return err
		}
	}
	return os.WriteFile(r.path+"/recalls", b.Bytes(), 0o644)
}

func (r *Recalls) Set(msg string) {
	recall := Recall{
		Time: time.Now(),
		Dir:  r.cwd,
		Msg:  msg,
	}
	if r.cwdIndex == -1 {
		r.recalls = append(r.recalls, recall)
		r.cwdIndex = len(r.recalls) - 1
	} else {
		r.recalls[r.cwdIndex] = recall
	}
}

func (r *Recalls) Msg() (string, error) {
	if r.cwdIndex == -1 {
		return "", errors.New("no recall set for this directory")
	}
	return r.recalls[r.cwdIndex].Msg, nil
}

func (r *Recalls) Path() string {
	return r.path
}

// makes sure there is no recall for cwd
func (r *Recalls) Delete() {
	if r.cwdIndex != -1 {
		r.recalls[r.cwdIndex] = r.recalls[len(r.recalls)-1]
		r.recalls = r.recalls[:len(r.recalls)-1]
	}
}
