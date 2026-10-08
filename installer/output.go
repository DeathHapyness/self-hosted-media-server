package installer

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	GREEN  = "\033[92m"
	RED    = "\033[91m"
	YELLOW = "\033[93m"
	CYAN   = "\033[96m"
	BOLD   = "\033[1m"
	RESET  = "\033[0m"
)

var UseColor = term.IsTerminal(int(os.Stdout.Fd()))

// func para posicionar frases no terminal.
func ncenter(width int, s string) string {
	n := (width - utf8.RuneCountInString(s)) / 2
	if n < 0 {
		n = 0
	}
	return strings.Repeat(" ", n) + s
}

func ColorText(text string, color string) string {
	if !UseColor {
		return text
	}
	return color + text + RESET
}

func PrintHeader(text string) {
	line := strings.Repeat("=", 40)

	fmt.Printf("\n%s\n", ColorText(line, CYAN))
	fmt.Println(ColorText(ncenter(40, text), CYAN+BOLD))
	fmt.Printf("%s\n\n", ColorText(line, CYAN))
}

func PrintSection(text string) {
	fmt.Printf("\n%s\n", ColorText(text, BOLD))
}

func PrintOk(text string) {
	fmt.Printf("%s %s\n", ColorText("[✓]", GREEN), text)
}

func PrintFail(text string) {
	fmt.Printf("%s %s\n", ColorText("[✗]", RED), text)
	log.Print(text)
}

func PrintWarning(text string) {
	fmt.Printf("%s %s\n", ColorText("[!]", YELLOW), text)
}

func PrintInfo(text string) {
	fmt.Printf("    %s\n", text)
}

func PrintDanger(text string) {
	fmt.Printf("%s %s\n", ColorText("[PERIGO]", RED+BOLD), text)
}
