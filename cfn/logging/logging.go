/*
Package logging provides support for logging to CloudWatch
within resource providers.
*/
package logging

import (
	"io"
	"log"
	"os"
	"syscall"
)

// stdErr is a reference to stderr that survives any os.Stderr redirections.
var stdErr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")

var providerLogOutput io.Writer

// SetProviderLogOutput configures the provider log output writer.
// When set, log messages go to both stderr and the provider writer (CloudWatch Logs).
func SetProviderLogOutput(w io.Writer) {
	log.SetOutput(w)
	providerLogOutput = w
}

// New sets up a logger that writes to stderr and (if configured) the provider log output.
func New(prefix string) *log.Logger {
	var w io.Writer

	if providerLogOutput != nil {
		w = io.MultiWriter(stdErr, providerLogOutput)
	} else {
		w = stdErr
	}

	return log.New(w, prefix, log.LstdFlags)
}
