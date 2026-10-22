package installer

import (
	"bufio"
	"fmt"
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

var stdin = bufio.NewReader(os.Stdin)

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

type Terminal struct{}

func (Terminal) Cabecalho(text string) {
	line := strings.Repeat("=", 40)

	fmt.Printf("\n%s\n", ColorText(line, CYAN))
	fmt.Println(ColorText(ncenter(40, text), CYAN+BOLD))
	fmt.Printf("%s\n\n", ColorText(line, CYAN))
}

func (Terminal) Secao(text string) {
	fmt.Printf("\n%s\n", ColorText(text, BOLD))
}

func (Terminal) Ok(text string) {
	fmt.Printf("%s %s\n", ColorText("[✓]", GREEN), text)
}

func (Terminal) Falha(text string) {
	fmt.Printf("%s %s\n", ColorText("[✗]", RED), text)
}

func (Terminal) Aviso(text string) {
	fmt.Printf("%s %s\n", ColorText("[!]", YELLOW), text)
}

func (Terminal) Info(text string) {
	fmt.Printf("    %s\n", text)
}

func (Terminal) Perigo(text string) {
	fmt.Printf("%s %s\n", ColorText("[PERIGO]", RED+BOLD), text)
}

func (Terminal) Texto(text string) {
	fmt.Print(text)
}

func (Terminal) Perguntar(prompt string) (string, error) {
	fmt.Print(prompt)
	linha, err := stdin.ReadString('\n')
	if err != nil && linha == "" {
		fmt.Println()
		return "", err
	}
	return strings.TrimSpace(linha), nil
}
