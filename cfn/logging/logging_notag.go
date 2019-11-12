// +build !logging

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

// SetProviderLogOutput ...
func SetProviderLogOutput(w io.Writer) {
	// no-op
}

// New sets up a logger that writes to the stderr
func New(prefix string) *log.Logger {
	// we create our own stderr since we're going to nuke the existing one
	return log.New(os.Stderr, prefix, log.LstdFlags)
}
