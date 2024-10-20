package recall

import (
	"fmt"
	"regexp"
)

var pathsRe = regexp.MustCompile(`f:\/?[^\/\s]+(?:\/[^\/\s]*)*`)

func (r Recall) Format() string {
	return pathsRe.ReplaceAllStringFunc(r.Msg, func(s string) string {
		path := s[2:]
		return fmt.Sprintf("\033]8;;file://%s\033\\%s\033]8;;\033\\", path, path)
	})
}
