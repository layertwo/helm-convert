package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/layertwo/helm-convert/cmd"
)

func main() {
	defer glog.Flush()

	if err := cmd.NewConvertCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
