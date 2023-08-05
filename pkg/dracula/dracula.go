package dracula

const FmtHex = "hex"
const FmtRgb = "rgb"
const FmtHsl = "hsl"

type Color interface {
	String() string
}
type ColorValues struct {
	Color
	Hex string
	RGB string
	HSL string
}

type Palette struct {
	Format                                                                                                  string
	Background, CurrentLine, Selection, Foreground, Comment, Cyan, Green, Orange, Pink, Purple, Red, Yellow Color
}

func New() *Palette {
	return &Palette{
		Format: "hex",
		Background: ColorValues{
			Hex: "#282a36",
			RGB: "40,42,54",
			HSL: "231 15% 18%",
		},
		CurrentLine: ColorValues{
			Hex: "#44475a",
			RGB: "68,71,90",
			HSL: "232 14% 31%",
		},
		Selection: ColorValues{
			Hex: "#44475a",
			RGB: "68,71,90",
			HSL: "232 14% 31%",
		},
		Foreground: ColorValues{
			Hex: "#f8f8f2",
			RGB: "248,248,242",
			HSL: "60 30% 96%",
		},
		Comment: ColorValues{
			Hex: "#6272a4",
			RGB: "98,114,164",
			HSL: "225 27% 51%",
		},
		Cyan: ColorValues{
			Hex: "#8be9fd",
			RGB: "139,233,253",
			HSL: "191 97% 77%",
		},
		Green: ColorValues{
			Hex: "#50fa7b",
			RGB: "80,250,123",
			HSL: "135 94% 65%",
		},
		Orange: ColorValues{
			Hex: "#ffb86c",
			RGB: "255,184,108",
			HSL: "31 100% 71%",
		},
		Pink: ColorValues{
			Hex: "#ff79c6",
			RGB: "255,121,198",
			HSL: "326 100% 74%",
		},
		Purple: ColorValues{
			Hex: "#bd93f9",
			RGB: "189,147,249",
			HSL: "265 89% 78%",
		},
		Red: ColorValues{
			Hex: "#ff5555",
			RGB: "255,85,85",
			HSL: "0 100% 67%",
		},
		Yellow: ColorValues{
			Hex: "#f1fa8c",
			RGB: "241,250,140",
			HSL: "65 92% 76%",
		},
	}
}

func (c ColorValues) String() string {
	switch dracula.Format {
	case "hex":
		return c.Hex
	case "rgb":
		return c.RGB
	case "hsl":
		return c.HSL
	default:
		return ""
	}
}

var dracula = New()

func (d *Palette) SetFormat(format string) {
	d.Format = format
}
