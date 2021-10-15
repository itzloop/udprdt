package utils

import (
	"fmt"
	"log"
	"os"
	"path"
)

//var logOnce = sync.Once{}
var fd = os.Stdout
var debug = false
var logger *log.Logger

func init() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("rdt.go: failed to get working directory")
		logger = log.Default()
		return
	}

	logName := fmt.Sprintf("host_%s.log", os.Args[1])

	fd, err = os.Create(path.Join(wd, logName))
	if err != nil {
		fmt.Println("rdt.go: failed to get working directory")
		logger = log.Default()
		return
	}

	if len(os.Args) > 3 && os.Args[3] == "debug"{
		debug = true
	}

	logger = log.New(fd, "", log.LstdFlags)
}



func Printf(format string, args ...interface{}) {
	if !debug {
		return
	}
	logger.Printf(format, args...)
}
