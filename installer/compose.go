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
		slog.Error("Could not list the repository", "path", root, "error", err)
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
	slog.Info("Services discovered", "total", len(encontrados), "services", strings.Join(nomesOrdenados(encontrados), ", "))
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
				portas[porta] = "unknown"
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
	slog.Info("Data folders eligible for removal", "total", len(unicos), "folders", strings.Join(unicos, ", "))
	return unicos
}

func escolherServicos(composeFiles map[string]string) map[string]string {
	nomes := nomesOrdenados(composeFiles)
	PrintSection(Tr("Quais serviços instalar?", "Which services to install?"))
	PrintInfo(Tr("0) Todos", "0) All"))
	for i, nome := range nomes {
		PrintInfo(fmt.Sprintf("%d) %s", i+1, nome))
	}

	escolha := readLine(Tr("\nNúmeros separados por vírgula (vazio = todos): ", "\nComma-separated numbers (empty = all): "))
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
		PrintWarning(Tr("Nenhum serviço válido selecionado — instalando todos.", "No valid service selected — installing all."))
		return composeFiles
	}
	return selecionados
}

type falhaServico struct {
	nome string
	erro string
}

func validateCompose(composeFiles map[string]string) error {
	PrintSection(Tr("Validando compose files...", "Validating compose files..."))
	var falhas []falhaServico
	for _, nome := range nomesOrdenados(composeFiles) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "config")
		if err != nil {
			falhas = append(falhas, falhaServico{nome, stderr})
		} else {
			// variável não definida é só aviso no compose, mas sobe o container com valor vazio
			aplicarCorrecoes(nome, composeFiles[nome], stderr)
			PrintOk(nome + Tr(" compose válido", " compose valid"))
		}
	}

	falhas = tentarCorrigir(falhas, composeFiles, func(nome string) (string, error) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "config")
		return stderr, err
	})

	if len(falhas) > 0 {
		for _, f := range falhas {
			PrintFail(f.nome + Tr(": configuração inválida", ": invalid configuration"))
			PrintInfo(f.erro)
		}
		return errors.New(Tr("Corrija os compose files acima e execute o instalador novamente.",
			"Fix the compose files above and run the installer again."))
	}
	return nil
}

func startServices(composeFiles map[string]string) error {
	PrintSection(Tr("Iniciando serviços...", "Starting services..."))
	cmd := strings.Join(composeBaseCmd(), " ")

	if DRY_RUN {
		for _, nome := range nomesOrdenados(composeFiles) {
			PrintInfo(fmt.Sprintf(Tr("[dry-run] rodaria: %s -f %s -p %s up -d", "[dry-run] would run: %s -f %s -p %s up -d"), cmd, composeFiles[nome], nome))
		}
		return nil
	}

	var falhas []falhaServico
	for _, nome := range nomesOrdenados(composeFiles) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "-p", nome, "up", "-d")
		if err != nil {
			falhas = append(falhas, falhaServico{nome, stderr})
		} else {
			PrintOk(nome + Tr(": container(s) iniciado(s)", " container(s) started"))
		}
	}

	falhas = tentarCorrigir(falhas, composeFiles, func(nome string) (string, error) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "-p", nome, "up", "-d")
		return stderr, err
	})

	if len(falhas) > 0 {
		for _, f := range falhas {
			PrintFail(f.nome + Tr(": falha ao iniciar", ": failed to start"))
			PrintInfo(f.erro)
		}
		return fmt.Errorf(Tr("Verifique os logs acima e '%s logs' de cada serviço para mais detalhes.",
			"Check the logs above and '%s logs' for each service for more details."), cmd)
	}
	return nil
}

func stopServices(composeFiles map[string]string) {
	PrintSection(Tr("Parando serviços...", "Stopping services..."))
	if len(composeFiles) == 0 {
		PrintWarning(Tr("Nenhum compose encontrado — nada para parar.", "No compose file found — nothing to stop."))
		return
	}

	cmd := strings.Join(composeBaseCmd(), " ")

	if DRY_RUN {
		for _, nome := range nomesOrdenados(composeFiles) {
			PrintInfo(fmt.Sprintf(Tr("[dry-run] rodaria: %s -f %s -p %s down", "[dry-run] would run: %s -f %s -p %s down"), cmd, composeFiles[nome], nome))
		}
		return
	}

	for _, nome := range nomesOrdenados(composeFiles) {
		_, stderr, err := runCompose("-f", composeFiles[nome], "-p", nome, "down")
		if err != nil {
			PrintWarning(nome + ": " + stderr)
		} else {
			PrintOk(nome + Tr(" parado", " stopped"))
		}
	}
}
