package main

import (
	"github.com/cloudradar-monitoring/tacoscript/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
