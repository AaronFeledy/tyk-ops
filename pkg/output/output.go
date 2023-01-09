package output

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

var (
	// outWriter is a writer defined by the user that replaces stdout
	outWriter io.Writer = os.Stdout
	// errWriter is a writer defined by the user that replaces stderr
	errWriter io.Writer = os.Stderr
	// Data is the customized logrus log used for stdout output.
	// Use this for data that can be piped to other applications.
	Data = log.New()
	// User is the customized logrus log used for stderr output.
	User = log.New()
	// OutFormatter is the formatter used for Data and User output
	OutFormatter = new(log.TextFormatter)
)

// init sets up loggers
func init() {
	Data.Out = outWriter
	User.Out = errWriter

	User.Formatter = OutFormatter

	OutFormatter.DisableTimestamp = true
	// Use default log.InfoLevel for Data
	Data.Level = log.InfoLevel // Data will by default always output
	logLevel := log.InfoLevel

	// Check environment variable for log level
	debugMode := os.Getenv("TYKOPS_DEBUG")
	if debugMode != "" {
		logLevel = log.DebugLevel
		Data.Level = log.DebugLevel
	}
	log.SetLevel(logLevel)
}
