// package recall handles storing and parsing recalls.
package recall

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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

// Parse parses a file of recalls.
func Parse() (Recalls, error) {
	cwd, err := fullCwd()
	if err != nil {
		return Recalls{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return Recalls{}, err
	}
	path := fmt.Sprintf("%s/.recall", home)
	recalls, cwdIndex, err := parseRecalls(path+"/recalls.json", cwd)
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

	var recalls []Recall
	if err = json.Unmarshal(f, &recalls); err != nil {
		return nil, 0, err
	}

	cwdIndex := -1
	for i, recall := range recalls {
		if recall.Dir == cwd {
			cwdIndex = i
			break
		}
	}

	return recalls, cwdIndex, nil
}

func (r *Recalls) Write() error {
	data, err := json.Marshal(r.recalls)
	if err != nil {
		return err
	}
	return os.WriteFile(r.path+"/recalls.json", data, 0o644)
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

func (r *Recalls) CwdHasRecall() bool {
	return r.cwdIndex != -1
}

func dirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

func fullCwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(cwd)
}
