package logging

import (
	"io"
	"log"
	"os"
	"syscall"
)

// define a new stdErr since we'll over-write the default stdout/err
// to prevent data leaking into the service account
var stdErr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")

var customerLogOutput io.Writer

const (
	loggerError = "Logger"
)

// SetCustomerLogOutput ...
func SetCustomerLogOutput(w io.Writer) {
	os.Stderr = nil
	os.Stdout = nil

	customerLogOutput = w
}

// New sets up a logger that writes to the stderr
func New(prefix string) *log.Logger {
	var w io.Writer

	if customerLogOutput != nil {
		w = io.MultiWriter(stdErr, customerLogOutput)
	} else {
		w = stdErr
	}

	// we create our own stderr since we're going to nuke the existing one
	return log.New(w, prefix, log.LstdFlags)
}
