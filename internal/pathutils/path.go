package pathutils

import (
	"os"
	"path/filepath"
	"regexp"
)

var fullLinuxPath = regexp.MustCompile("^(/[^/ ]*)+/?$")

func Valid(path string) bool {
	return fullLinuxPath.MatchString(path)
}

func DirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

// returns full cwd
func Cwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(cwd)
}
