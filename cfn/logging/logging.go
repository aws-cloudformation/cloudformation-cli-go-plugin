package logging

import (
	"io"
	"log"
	"os"
	"syscall"

	"github.com/aws-cloudformation/aws-cloudformation-rpdk-go-plugin/cfn/cfnerr"
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
func New(prefix string) (*log.Logger, error) {
	if customerLogOutput == nil {
		return nil, cfnerr.New(loggerError, "Customer log output not defined", nil)
	}

	// we create our own stderr since we're going to nuke the existing one
	return log.New(io.MultiWriter(stdErr, customerLogOutput), prefix, log.LstdFlags), nil
}
