package color

import (
	"fmt"
	"github.com/AaronFeledy/tyk-ops/pkg/dracula"
)

const FmtHex = dracula.FmtHex
const FmtRgb = dracula.FmtRgb
const FmtHsl = dracula.FmtHsl

var palette = dracula.New()

func Init() {
	SetFormat(FmtHex)
}

func SetFormat(format string) {
	palette.SetFormat(format)
}

var (
	Background  = fmt.Sprint(palette.Background)
	CurrentLine = fmt.Sprint(palette.CurrentLine)
	Selection   = fmt.Sprint(palette.Selection)
	Foreground  = fmt.Sprint(palette.Foreground)
	Comment     = fmt.Sprint(palette.Comment)
	Cyan        = fmt.Sprint(palette.Cyan)
	Green       = fmt.Sprint(palette.Green)
	Orange      = fmt.Sprint(palette.Orange)
	Pink        = fmt.Sprint(palette.Pink)
	Purple      = fmt.Sprint(palette.Purple)
	Red         = fmt.Sprint(palette.Red)
	Yellow      = fmt.Sprint(palette.Yellow)
)
