package installer

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	reServico = regexp.MustCompile(`^\s{2}([a-zA-Z0-9_-]+):\s*$`)
	rePorta   = regexp.MustCompile(`-\s*"?(\d+):\d+(?:/\w+)?"?`)
	reVolumes = regexp.MustCompile(`^volumes:\s*$`)
)

func repoRoot() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
	}
	dir, _ := os.Getwd()
	return dir
}

func resolver(caminho string) string {
	abs, err := filepath.Abs(caminho)
	if err != nil {
		return filepath.Clean(caminho)
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		return real
	}
	return abs
}

func isDir(caminho string) bool {
	info, err := os.Stat(caminho)
	return err == nil && info.IsDir()
}

func nomesOrdenados(m map[string]string) []string {
	return slices.Sorted(maps.Keys(m))
}

func composeEnv() []string {
	return append(os.Environ(), "DATA_DIR="+BASE_DIR, "MEDIA_DIR="+MEDIA_DIR)
}

func runCompose(args ...string) (stdout, stderr string, err error) {
	cmd := composeBaseCmd()
	return runEnv(composeEnv(), cmd[0], append(cmd[1:], args...)...)
}

func findComposeFiles() map[string]string {
	root := repoRoot()
	encontrados := map[string]string{}

	entradas, err := os.ReadDir(root)
	if err != nil {
		slog.Error("Nao foi possivel listar o repositorio", "caminho", root, "erro", err)
		return encontrados
	}
	for _, e := range entradas {
		nome := e.Name()
		if !isDir(filepath.Join(root, nome)) {
			continue
		}
		if strings.HasPrefix(nome, ".") || NON_SERVICE_DIRS[nome] {
			continue
		}
		for _, candidato := range COMPOSE_CANDIDATES {
			caminho := filepath.Join(root, nome, candidato)
			if _, err := os.Stat(caminho); err == nil {
				encontrados[nome] = caminho
				break
			}
		}
	}
	slog.Info("Servicos descobertos", "total", len(encontrados), "servicos", strings.Join(nomesOrdenados(encontrados), ", "))
	return encontrados
}

func extractPortsFromCompose(composePath string) map[int]string {
	portas := map[int]string{}
	dados, err := os.ReadFile(composePath)
	if err != nil {
		return portas
	}

	servicoAtual := ""
	for _, linha := range strings.Split(string(dados), "\n") {
		linha = strings.TrimRight(linha, "\r")
		if m := reServico.FindStringSubmatch(linha); m != nil {
			servicoAtual = m[1]
			continue
		}
		if m := rePorta.FindStringSubmatch(linha); m != nil {
			porta, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			if servicoAtual == "" {
				portas[porta] = "desconhecido"
			} else {
				portas[porta] = servicoAtual
			}
		}
	}
	return portas
}

func extractDataDirsFromCompose(composePath string) []string {
	root := resolver(repoRoot())
	serviceDir := resolver(filepath.Dir(composePath))
	protegidos := map[string]bool{}
	for _, p := range PROTECTED_REPO_DIRS {
		protegidos[resolver(filepath.Join(root, p))] = true
	}

	dados, err := os.ReadFile(composePath)
	if err != nil {
		return nil
	}

	var dirs []string
	emVolumes := false
	volumesIndent := 0

	for _, linha := range strings.Split(string(dados), "\n") {
		linha = strings.TrimRight(linha, "\r")
		limpa := strings.TrimSpace(linha)
		if limpa == "" || strings.HasPrefix(limpa, "#") {
			continue
		}
		indent := len(linha) - len(strings.TrimLeft(linha, " \t"))

		if reVolumes.MatchString(limpa) {
			emVolumes = true
			volumesIndent = indent
			continue
		}
		if !emVolumes {
			continue
		}
		if !strings.HasPrefix(limpa, "- ") {
			if indent <= volumesIndent {
				emVolumes = false
			}
			continue
		}

		host, _, _ := strings.Cut(limpa[2:], ":")
		host = strings.TrimSpace(host)
		if strings.Contains(host, "$") || !strings.HasPrefix(host, "./") {
			continue
		}

		alvo := resolver(filepath.Join(serviceDir, host[2:]))
		if alvo == serviceDir || alvo == root || protegidos[alvo] {
			continue
		}
		if !strings.HasPrefix(alvo, root+string(filepath.Separator)) || !isDir(alvo) {
			continue
		}
		dirs = append(dirs, alvo)
	}
	return dirs
}

func collectDataDirs(composeFiles map[string]string) []string {
	var alvos []string
	if isDir(BASE_DIR) {
		alvos = append(alvos, resolver(BASE_DIR))
	}
	for _, composePath := range composeFiles {
		alvos = append(alvos, extractDataDirsFromCompose(composePath)...)
	}

	slices.Sort(alvos)
	alvos = slices.Compact(alvos)

	var unicos []string
	for _, alvo := range alvos {
		dentroDeOutro := slices.ContainsFunc(unicos, func(pai string) bool {
			return strings.HasPrefix(alvo, pai+string(filepath.Separator))
		})
		if !dentroDeOutro {
			unicos = append(unicos, alvo)
		}
	}
	slog.Info("Pastas de dados candidatas a remocao", "total", len(unicos), "pastas", strings.Join(unicos, ", "))
	return unicos
}

func escolherServicos(composeFiles map[string]string) map[string]string {
	nomes := nomesOrdenados(composeFiles)
	PrintSection("Quais serviços instalar?")
	PrintInfo("0) Todos")
	for i, nome := range nomes {
		PrintInfo(fmt.Sprintf("%d) %s", i+1, nome))
	}

	escolha := readLine("\nNúmeros separados por vírgula (vazio = todos): ")
	if escolha == "" || escolha == "0" {
		return composeFiles
	}

	selecionados := map[string]string{}
	for _, parte := range strings.Split(escolha, ",") {
		parte = strings.TrimSpace(parte)
		n, err := strconv.Atoi(parte)
		if err != nil || strings.ContainsAny(parte, "+-") {
			continue
		}
		if n >= 1 && n <= len(nomes) {
			nome := nomes[n-1]
			selecionados[nome] = composeFiles[nome]
		}
	}

	if len(selecionados) == 0 {
		PrintWarning("Nenhum serviço válido selecionado — instalando todos.")
		return composeFiles
	}
	return selecionados
}

type falhaServico struct {
	nome string
	erro string
}

func validateCompose(composeFiles map[string]string) error {
	PrintSection("Validating compose files...")
	var falhas []falhaServico
	for _, nome := range nomesOrdenados(composeFiles) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "config")
		if err != nil {
			falhas = append(falhas, falhaServico{nome, stderr})
		} else {
			PrintOk(nome + " compose válido")
		}
	}

	falhas = tentarCorrigir(falhas, composeFiles, func(nome string) (string, error) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "config")
		return stderr, err
	})

	if len(falhas) > 0 {
		for _, f := range falhas {
			PrintFail(f.nome + ": configuração inválida")
			PrintInfo(f.erro)
		}
		return errors.New("Corrija os compose files acima e execute o instalador novamente.")
	}
	return nil
}

func startServices(composeFiles map[string]string) error {
	PrintSection("Starting services...")
	cmd := strings.Join(composeBaseCmd(), " ")

	if DRY_RUN {
		for _, nome := range nomesOrdenados(composeFiles) {
			PrintInfo(fmt.Sprintf("[dry-run] rodaria: %s -f %s -p %s up -d", cmd, composeFiles[nome], nome))
		}
		return nil
	}

	var falhas []falhaServico
	for _, nome := range nomesOrdenados(composeFiles) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "-p", nome, "up", "-d")
		if err != nil {
			falhas = append(falhas, falhaServico{nome, stderr})
		} else {
			PrintOk(nome + " container(s) started")
		}
	}

	falhas = tentarCorrigir(falhas, composeFiles, func(nome string) (string, error) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "-p", nome, "up", "-d")
		return stderr, err
	})

	if len(falhas) > 0 {
		for _, f := range falhas {
			PrintFail(f.nome + ": falha ao iniciar")
			PrintInfo(f.erro)
		}
		return fmt.Errorf("Verifique os logs acima e '%s logs' de cada serviço para mais detalhes.", cmd)
	}
	return nil
}

func stopServices(composeFiles map[string]string) {
	PrintSection("Parando serviços...")
	if len(composeFiles) == 0 {
		PrintWarning("Nenhum compose encontrado — nada para parar.")
		return
	}

	cmd := strings.Join(composeBaseCmd(), " ")

	if DRY_RUN {
		for _, nome := range nomesOrdenados(composeFiles) {
			PrintInfo(fmt.Sprintf("[dry-run] rodaria: %s -f %s -p %s down", cmd, composeFiles[nome], nome))
		}
		return
	}

	for _, nome := range nomesOrdenados(composeFiles) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "-p", nome, "down")
		if err != nil {
			PrintWarning(nome + ": " + stderr)
		} else {
			PrintOk(nome + " parado")
		}
	}
}
