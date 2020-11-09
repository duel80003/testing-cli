package main

import (
	"twilio-cli/task"
)

func main() {
	task.PrepareData()
	test := task.NewInstance()
	test.Start()
}
