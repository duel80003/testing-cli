package printer

import (
	"fmt"

	"github.com/buger/goterm"
	"github.com/fatih/color"
)

var cyan = color.New(color.FgCyan).SprintFunc()
var hiRed = color.New(color.FgHiRed).SprintFunc()
var white = color.New(color.FgHiWhite).SprintFunc()
var yellow = color.New(color.FgHiYellow).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var divider = color.New(color.Bold, color.FgYellow).PrintlnFunc()
var printHTTPInfo = color.New(color.FgHiWhite)

// Info for print info message
func Info(s interface{}) {
	fmt.Printf("%-15s %v \n", cyan("INFO "), white(s))
}

// Error for print error message
func Error(s interface{}) {
	// hiRed("Error: %v \n", s)
	fmt.Printf("%-15s %v \n", hiRed("Error "), white(s))
}

// Print is fmt.Println()
func Print(s interface{}) {
	fmt.Println(s)
}

// Divider is used to separate sections of conversation
func Divider() {
	width := goterm.Width() - 1
	d := ""
	for i := 0; i < width; i++ {
		d += "-"
	}
	divider(d)
}

// ShowAnswer print the user input
func ShowAnswer(s interface{}) {
	printHTTPInfo.Printf("%-18s", "Send to Twilio:")
	fmt.Printf("%v \n", blue(s))
}

// ShowQuestion print the response form twilio
func ShowQuestion(s interface{}) {
	printHTTPInfo.Printf("%-18s", "Next Question:")
	fmt.Printf("%v \n", red(s))
}

// ColorInstance return a color instance
func ColorInstance(c string) *color.Color {
	switch c {
	case "yellow":
		return color.New(color.FgHiYellow)
	case "blue":
		return color.New(color.FgBlue)
	case "red":
		return color.New(color.FgRed)
	case "magenta":
		return color.New(color.FgMagenta)
	case "cyan":
		return color.New(color.FgCyan)
	default:
		return color.New(color.FgHiWhite)
	}
}
