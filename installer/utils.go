package installer

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func run(name string, args ...string) (stdout, stderr string, err error) {
	return runEnv(nil, name, args...)
}

func runEnv(env []string, name string, args ...string) (stdout, stderr string, err error) {
	comando := name + " " + strings.Join(args, " ")
	slog.Info("Running", "command", comando)

	var out, errOut bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()

	stdout = out.String()
	stderr = strings.TrimSpace(errOut.String())
	if err != nil {
		slog.Error("Command failed", "command", comando, "error", err, "stderr", stderr)
	} else if s := strings.TrimSpace(stdout); s != "" {
		slog.Debug("stdout", "command", comando, "output", s)
	}
	return stdout, stderr, err
}

func confirmDangerous(explicacao, palavra string) bool {
	PrintDanger(explicacao)
	PrintInfo(fmt.Sprintf(Tr("Isto NÃO pode ser desfeito automaticamente. Digite exatamente \"%s\" para continuar.",
		"This can NOT be undone automatically. Type exactly \"%s\" to continue."), palavra))
	confirmado := readLine("> ") == palavra
	if confirmado {
		slog.Warn("Destructive action CONFIRMED by the user", "explanation", explicacao)
	} else {
		slog.Info("Destructive action cancelled by the user")
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
	readLine(Tr("\nPressione Enter para voltar ao menu...", "\nPress Enter to return to the menu..."))
}
