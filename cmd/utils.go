package cmd

import (
	"io/ioutil"
	"os"
	"strings"
	"twilio-test-cli/logger"
)

const path = "./"

//ConcatString return a concat string
func ConcatString(args []string) string {
	return strings.Join(args, "")
}
func createFilePath(args []string) string {
	fileName := ConcatString(args)
	logger.Info("Reading config file: " + yellow(fileName))
	return ConcatString([]string{path, fileName})
}

func readJSONFile(fullPath string) []byte {
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		logger.Error(ConcatString([]string{"Read file", fullPath, " failure."}))
		os.Exit(1)
	}
	return data
}
