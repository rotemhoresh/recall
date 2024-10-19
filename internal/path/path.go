package path

import (
	"os"
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
