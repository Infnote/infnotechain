package utils

import (
	"regexp"
	"strings"
)

func Intend(target string, level int) string {
	return regexp.MustCompile("(?m)^").ReplaceAllString(target, strings.Repeat("\t", level))
}
