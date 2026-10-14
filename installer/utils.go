package installer

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var stdin = bufio.NewReader(os.Stdin)

func run(name string, args ...string) (stdout, stderr string, err error) {
	return runEnv(nil, name, args...)
}

// runEnv é o run(cmd, env=...) do Python: env substitui o ambiente inteiro do
// comando (use append(os.Environ(), "CHAVE=valor") para herdar o atual).
// env nil herda o ambiente do processo.
func runEnv(env []string, name string, args ...string) (stdout, stderr string, err error) {
	comando := name + " " + strings.Join(args, " ")
	slog.Info("Executando", "comando", comando)

	var out, errOut bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Env = env
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
	if err != nil && linha == "" {
		fmt.Println()
		slog.Warn("Entrada encerrada; saindo", "erro", err)
		FinalizarExecucao("cancelada (entrada encerrada)")
		os.Exit(1)
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

func tamanhoLegivel(caminho string) string {
	var total int64
	filepath.WalkDir(caminho, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		// os.Stat segue symlinks, igual ao os.path.getsize
		info, err := os.Stat(p)
		if err != nil || info.IsDir() {
			return nil
		}
		total += info.Size()
		return nil
	})

	tamanho := float64(total)
	for _, unidade := range []string{"B", "KB", "MB", "GB", "TB"} {
		if tamanho < 1024 {
			return fmt.Sprintf("%.0f%s", tamanho, unidade)
		}
		tamanho /= 1024
	}
	return fmt.Sprintf("%.0fPB", tamanho)
}

func pressEnterToContinue() {
	readLine("\nPressione Enter para voltar ao menu...")
}
