package installer

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"time"
)

func CheckOS() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf(Tr("Sistema operacional não suportado: %s. Este instalador requer Linux.",
			"Unsupported operating system: %s. This installer requires Linux."), runtime.GOOS)
	}
	slog.Info("Operating system", "os", runtime.GOOS)
	PrintOk(Tr("Linux detectado", "Linux detected"))
	return nil
}

func CheckPrivileges() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf(Tr("Este instalador precisa ser executado como root (sudo).\n    Execute: sudo %s",
			"This installer must be run as root (sudo).\n    Run: sudo %s"), os.Args[0])
	}
	PrintOk(Tr("Executando com privilégios suficientes", "Running with sufficient privileges"))
	return nil
}

func isMount(caminho string) bool {
	var st, pai syscall.Stat_t
	if err := syscall.Lstat(caminho, &st); err != nil {
		return false
	}
	if st.Mode&syscall.S_IFMT == syscall.S_IFLNK {
		return false
	}
	if err := syscall.Lstat(filepath.Join(caminho, ".."), &pai); err != nil {
		return false
	}
	return st.Dev != pai.Dev || st.Ino == pai.Ino
}

func CheckMediaMount(caminho string) bool {
	if _, err := os.Stat(caminho); err != nil {
		return false
	}
	if !isMount(caminho) {
		return false
	}
	PrintOk(fmt.Sprintf(Tr("%s montado", "%s mounted"), caminho))
	return true
}

func resolveMediaMountInteractive() bool {
	if CheckMediaMount(MEDIA_DIR) {
		return true
	}
	PrintWarning(fmt.Sprintf(Tr(
		"%s não existe ou não é um mount separado. Ter um disco dedicado "+
			"é opcional — você pode seguir usando esse caminho como uma pasta comum no disco raiz.",
		"%s does not exist or is not a separate mount. A dedicated disk "+
			"is optional — you can keep using this path as a regular folder on the root disk."),
		MEDIA_DIR))
	return Confirmou(readLine(Tr("Continuar mesmo assim? [s/N]: ", "Continue anyway? [y/N]: ")), false)
}

func primeiroExistente(caminho string) string {
	for {
		if _, err := os.Stat(caminho); err == nil {
			return caminho
		}
		pai := filepath.Dir(caminho)
		if pai == caminho {
			return caminho
		}
		caminho = pai
	}
}

func espacoLivreGB(caminho string) (float64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(caminho, &stat); err != nil {
		return 0, err
	}
	return float64(stat.Bavail*uint64(stat.Bsize)) / (1024 * 1024 * 1024), nil
}

func checkDiskSpace() error {
	dataPath := primeiroExistente(BASE_DIR)
	dataLivre, err := espacoLivreGB(dataPath)
	if err != nil {
		return fmt.Errorf(Tr("Não foi possível verificar o espaço em %s: %v", "Could not check free space on %s: %v"), dataPath, err)
	}
	if dataLivre < MIN_DISK_SPACE_DATA_GB {
		return fmt.Errorf(Tr("Espaço insuficiente em %s: %.1fGB livres, mínimo exigido %dGB.",
			"Not enough space on %s: %.1fGB free, at least %dGB required."),
			dataPath, dataLivre, MIN_DISK_SPACE_DATA_GB)
	}
	slog.Info("Free space", "path", dataPath, "gb", fmt.Sprintf("%.1f", dataLivre))

	mediaPath := primeiroExistente(MEDIA_DIR)
	mediaLivre, err := espacoLivreGB(mediaPath)
	if err != nil {
		return fmt.Errorf(Tr("Não foi possível verificar o espaço em %s: %v", "Could not check free space on %s: %v"), mediaPath, err)
	}
	slog.Info("Free space", "path", mediaPath, "gb", fmt.Sprintf("%.1f", mediaLivre))
	if mediaLivre < MIN_DISK_SPACE_MEDIA_GB {
		PrintWarning(fmt.Sprintf(Tr("Pouco espaço livre em %s: %.1fGB (recomendado: %dGB+). Continuando mesmo assim.",
			"Low free space on %s: %.1fGB (recommended: %dGB+). Continuing anyway."),
			mediaPath, mediaLivre, MIN_DISK_SPACE_MEDIA_GB))
	} else {
		PrintOk(fmt.Sprintf(Tr("Armazenamento disponível (%.1fGB livres em %s)", "Storage available (%.1fGB free on %s)"), mediaLivre, mediaPath))
	}
	return nil
}

func checkPorts(composeFiles map[string]string) error {
	esperadas := map[int]string{}
	for servico, composePath := range composeFiles {
		for porta, container := range extractPortsFromCompose(composePath) {
			esperadas[porta] = servico + "/" + container
		}
	}

	if len(esperadas) == 0 {
		PrintWarning(Tr("Não foi possível extrair portas dos compose files, usando lista padrão.",
			"Could not read ports from the compose files, using the default list."))
		for porta, servico := range DEFAULT_EXPECTED_PORTS {
			esperadas[porta] = servico
		}
	}

	portas := make([]int, 0, len(esperadas))
	for porta := range esperadas {
		portas = append(portas, porta)
	}
	slices.Sort(portas)

	var ocupadas []int
	for _, porta := range portas {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", porta), 500*time.Millisecond)
		if err == nil {
			conn.Close()
			ocupadas = append(ocupadas, porta)
		}
	}

	if len(ocupadas) > 0 {
		var lista []string
		for _, porta := range ocupadas {
			lista = append(lista, fmt.Sprintf("%d (%s)", porta, esperadas[porta]))
		}
		slog.Warn("Ports in use", "ports", strings.Join(lista, ", "))
		PrintFail(Tr("Portas em uso detectadas:", "Ports already in use:"))
		for _, porta := range ocupadas {
			PrintInfo(fmt.Sprintf(Tr("porta %d (%s) já está em uso", "port %d (%s) is already in use"), porta, esperadas[porta]))
		}
		return errors.New(Tr("Libere as portas acima (pare o serviço que as está usando) ou ajuste "+
			"o docker-compose.yml antes de continuar.",
			"Free the ports above (stop the service using them) or adjust "+
				"the docker-compose.yml before continuing."))
	}
	PrintOk(Tr("Portas necessárias disponíveis", "Required ports available"))
	return nil
}
