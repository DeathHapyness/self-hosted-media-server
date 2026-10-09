package installer

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

var stdin = bufio.NewReader(os.Stdin)

func run(name string, args ...string) (stdout, stderr string, err error) {
	comando := name + " " + strings.Join(args, " ")
	slog.Info("Executando", "comando", comando)

	var out, errOut bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()

	stdout = out.String()
	stderr = strings.TrimSpace(errOut.String())
	if err != nil {
		slog.Error("Comando falhou", "comando", comando, "erro", err, "stderr", stderr)
	} else if s := strings.TrimSpace(stdout); s != "" {
		slog.Debug("stdout", "comando", comando, "saida", s)
	}
	return stdout, stderr, err
}

func readLine(prompt string) string {
	fmt.Print(prompt)
	linha, err := stdin.ReadString('\n')
	if err != nil {
		fmt.Println()
	}
	return strings.TrimSpace(linha)
}

func confirmDangerous(explicacao, palavra string) bool {
	PrintDanger(explicacao)
	PrintInfo(fmt.Sprintf("Isto NÃO pode ser desfeito automaticamente. Digite exatamente \"%s\" para continuar.", palavra))
	confirmado := readLine("> ") == palavra
	if confirmado {
		slog.Warn("Ação destrutiva CONFIRMADA pelo usuário", "explicacao", explicacao)
	} else {
		slog.Info("Ação destrutiva cancelada pelo usuário")
	}
	return confirmado
}

func pressEnterToContinue() {
	readLine("\nPressione Enter para voltar ao menu...")
}
