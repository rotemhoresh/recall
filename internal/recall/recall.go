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
	"fmt"
	"os"
	"strings"
	"time"

	pathutils "github.com/rotemhoresh/recall/internal/path"
)

const Separator = " ! "

type Recall struct {
	Time time.Time
	Dir  string
	Msg  string
}

type Recalls []Recall

type ParsingError string

func (e ParsingError) Error() string {
	return fmt.Sprintf("recall: parsing error: %s", string(e))
}

// Parse parses a file of recalls.
func Parse(path string) (Recalls, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(f), "\n")
	lines = lines[:len(lines)-1] // last one is always empty
	// if the file is in a valid format, each line accounts for a recall
	recalls := make(Recalls, 0, len(lines))

	for _, ln := range lines {
		segments := strings.Split(ln, Separator)
		if l := len(segments); l != 3 {
			return nil, ParsingError(fmt.Sprintf("expected line to have %d segments, but got %d", 3, l))
		}
		t, err := time.Parse(time.RFC3339, segments[0])
		if err != nil {
			return nil, ParsingError(err.Error())
		}
		if !pathutils.Valid(segments[1]) {
			return nil, ParsingError(fmt.Sprintf("invalid path format: %s", segments[1]))
		}
		if dirExists, err := pathutils.DirExists(segments[1]); err != nil {
			return nil, err
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
	}

	return recalls, nil
}

func (l Recalls) Write(path string) error {
	var b bytes.Buffer
	for _, recall := range l {
		if _, err := b.WriteString(fmt.Sprintf("%s ! %s ! %s\n", recall.Time.Format(time.RFC3339), recall.Dir, recall.Msg)); err != nil {
			return err
		}
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}
