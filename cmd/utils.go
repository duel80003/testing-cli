package cmd

import (
	"strings"
)

//ConcatString return a concat string
func ConcatString(args []string) string {
	return strings.Join(args, "")
}
