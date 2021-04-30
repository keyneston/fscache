package main

import (
	"flag"
	"log"
	"syscall"

	"github.com/keyneston/fscachemonitor/fscache"
	"github.com/sirupsen/logrus"
)

func main() {
	var filename string
	var root string
	var debug bool

	flag.StringVar(&root, "r", "", "Root directory to monitor")
	flag.StringVar(&filename, "f", "", "File to output to")
	flag.BoolVar(&debug, "debug", false, "Set debug logging")
	flag.Parse()

	logrus.SetLevel(logrus.ErrorLevel)
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if root == "" {
		log.Fatalf("Must specify root to watch")
	}
	if filename == "" {
		log.Fatalf("Must specify file to output cache to")
	}

	if err := setLimits(); err != nil {
		log.Fatalf("Error updating limits: %v", err)
	}

	fs, err := fscache.New(filename, root)
	if err != nil {
		log.Fatalf("Error starting monitor: %v", err)
	}
	fs.Run()

}

func setLimits() error {
	var limit syscall.Rlimit

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}

	limit.Cur = 9999999
	limit.Max = 9999999

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}

	log.Printf("Limits: %v", limit)

	return nil
}
