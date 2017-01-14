package refcode

import (
	"io/ioutil"
	"log"
	"os"
)

// Verbose exports verbose logger.
var Verbose *log.Logger

// ErrorLog exports error logger.
var ErrorLog *log.Logger

func init() {
	Verbose = log.New(ioutil.Discard, "[verbose] ", 0)
	ErrorLog = log.New(os.Stderr, "Error: ", 0)
}

// EnableVerboseLog enables verbose log functionality.
func EnableVerboseLog() {
	Verbose.SetFlags(0)
	Verbose.SetOutput(os.Stdout)
}
