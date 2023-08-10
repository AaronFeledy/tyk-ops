package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
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
	// UserFormatter is the formatter used for User output
	UserFormatter = new(PlainFormatter)
	// DebugOutput is a global output object that can be used to print debug messages
	DebugOutput = NewOutput(outWriter, errWriter)
)

// init sets up loggers
func init() {
	Data.Out = outWriter
	User.Out = errWriter

	User.Formatter = UserFormatter

	DataFormatter := new(PlainFormatter)
	Data.Formatter = DataFormatter
	// Use default log.InfoLevel for Data
	Data.Level = log.InfoLevel // Data will by default always output
	logLevel := log.InfoLevel

	// Check environment variable for log level
	debugMode := os.Getenv("TYKOPS_DEBUG")
	if debugMode != "" {
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)
}

// Output is a struct that can be used to print data and flair to stdout and stderr.
type Output struct {
	// outWriter is the writer to use for stdout. nil will use os.Stdout
	outWriter io.Writer
	// errWriter is the writer to use for stderr. nil will use os.Stderr
	errWriter io.Writer
}

// NewOutput creates a new output object. If stdout or stderr are nil, os.Stdout or os.Stderr will be used.
func NewOutput(out io.Writer, err io.Writer) *Output {
	if out == nil {
		out = outWriter
	}
	if err == nil {
		err = errWriter
	}
	return &Output{
		outWriter: out,
		errWriter: err,
	}
}

// NewFromCmd returns an Output object that uses the stdout and stderr of the supplied command.
func NewFromCmd(cmd *cobra.Command) *Output {
	out := cmd.OutOrStdout()
	err := cmd.ErrOrStderr()
	return NewOutput(out, err)
}

// Dataln prints the data to the stdout writer followed by a line break to the stderr writer. This allows the data to be
// piped to other commands without including the line break that would be printed to the console.
func (o *Output) Dataln(data string) {
	outMsg := &FlairedData{
		data:   data,
		output: o,
	}
	outMsg.Post("\n").Print()
}

// Dataf uses printf style formatting to print the data to the stdout writer.
func (o *Output) Dataf(format string, args ...interface{}) {
	_, err := fmt.Fprintf(o.outWriter, format, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Printf(format, args...)
	}
}

// Infoln prints an informational message to the stderr writer followed by a line break.
func (o *Output) Infoln(msg string) {
	o.Infof("%s\n", msg)
}

// Infof uses printf style formatting to print an informational message to the stderr writer.
func (o *Output) Infof(format string, args ...interface{}) {
	_, err := fmt.Fprintf(o.errWriter, format, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// DataWithFlair Prints only the data to stdout so it can be piped to other commands without including
// the surrounding text.
// Use Pre() and Post() methods to add surrounding text that will be displayed to the user but will
// not be included in the piped output.
func (o *Output) DataWithFlair(data string) *FlairedData {
	return &FlairedData{
		data:   data,
		output: o,
	}
}

// PrettyString prints a string as pretty JSON to the stdout writer.
func (o *Output) PrettyString(str string) {
	var prettyJSON bytes.Buffer
	str = strings.TrimSpace(str)
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err == nil {
		o.Dataf("%s", prettyJSON.String())
	} else {
		// Trim the string to remove any leading or trailing whitespace
		str = strings.TrimSpace(str)
		o.Dataf("%s", str)
	}
	o.Infof("\n")
}

// Debug prints the message to the stderr writer if debug mode is enabled.
func (o *Output) Debug(msg string) {
	Debug(msg)
}

// Debug prints the message to the stderr writer if debug mode is enabled.
func Debug(msg string) {
	if Data.Level == log.DebugLevel {
		DebugOutput.Infof("%s\n", msg)
	}
}

// FlairedData is a struct that can be used to print data with surrounding flair.
type FlairedData struct {
	output    *Output
	data      string
	flairPre  string
	flairPost string
}

// Pre sets the flair to be printed before the data.
func (f *FlairedData) Pre(flair string) *FlairedData {
	f.flairPre = flair
	return f
}

// Post sets the flair to be printed after the data.
func (f *FlairedData) Post(flair string) *FlairedData {
	f.flairPost = flair
	return f
}

// Print will output the flaired message
func (f *FlairedData) Print() {
	if f.output == nil {
		f.output = NewOutput(Data.Out, User.Out)
	}
	out := f.output

	// Print only the data to stdout so it can be piped to other commands without including the surrounding text.
	out.Infof("%s", f.flairPre)
	out.Dataf("%s", f.data)

	// Rewrite the previous message to stderr so that it is still shown to the user if it's been captured.
	preSuffix := ""
	colorUnformat := fmt.Sprintf("%s[%dm", "\x1b", color.Reset)
	for {
		// Remove formatting characters that might follow a line break character
		if strings.HasSuffix(f.flairPre, colorUnformat) {
			f.flairPre = strings.TrimSuffix(f.flairPre, colorUnformat)
			preSuffix += colorUnformat
		} else {
			break
		}
	}
	runes := bytes.Runes([]byte(f.flairPre))
	f.flairPre = f.flairPre + preSuffix
	if len(runes) > 0 && runes[len(runes)-1] == '\n' {
		out.Infof("\r%s", f.data)
	} else {
		// When data and prefix are on the same line, we need to reprint both
		out.Infof("\r%s%s", f.flairPre, f.data)
	}

	out.Infof("%s", f.flairPost)
}

// Println will output the flaired message followed by a line break
func (f *FlairedData) Println() {
	f.Print()
	f.output.Infoln("")
}

// PlainFormatter defines a output formatter with no special frills
type PlainFormatter struct{}

// Format is called by logrus to format the output. We just return the string as is.
func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	switch entry.Level {
	case log.DebugLevel, log.TraceLevel:
		entry.Message = fmt.Sprintf("DEBUG  %v", entry.Message)
	case log.WarnLevel:
		entry.Message = fmt.Sprintf("%s  %s\n", color.YellowString("WARN "), entry.Message)
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		entry.Message = fmt.Sprintf("%s  %s\n", color.RedString("ERROR"), entry.Message)
	case log.InfoLevel:
	default:
	}
	return []byte(entry.Message), nil
}
