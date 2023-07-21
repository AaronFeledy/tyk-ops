package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
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

// DataWithFlair Prints only the data to stdout so it can be piped to other commands without including
// the surrounding text.
// Use Pre() and Post() methods to add surrounding text that will be displayed to the user but will
// not be included in the piped output.
func DataWithFlair(data string) *FlairedData {
	return &FlairedData{
		data: data,
	}
}

func PrettyString(str string) {
	var prettyJSON bytes.Buffer
	str = strings.TrimSpace(str)
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err == nil {
		Data.Print(prettyJSON.String())
	} else {
		// Trim the string to remove any leading or trailing whitespace
		str = strings.TrimSpace(str)
		Data.Print(str)
	}
	User.Printf("\n")
}

// FlairedData is a struct that can be used to print data with surrounding flair.
type FlairedData struct {
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
	// Print only the data to stdout so it can be piped to other commands without including the surrounding text.
	User.Printf("%s", f.flairPre)
	Data.Printf("%s", f.data)

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
		User.Printf("\r%s", f.data)
	} else {
		// When data and prefix are on the same line, we need to reprint both
		User.Printf("\r%s%s", f.flairPre, f.data)
	}

	User.Printf("%s", f.flairPost)
}

// Println will output the flaired message followed by a line break
func (f *FlairedData) Println() {
	f.Print()
	User.Println()
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
