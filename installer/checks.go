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
		return fmt.Errorf("Sistema operacional não suportado: %s. Este instalador requer Linux.", runtime.GOOS)
	}
	slog.Info("Sistema operacional", "so", runtime.GOOS)
	PrintOk("Linux detected")
	return nil
}

func CheckPrivileges() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("Este instalador precisa ser executado como root (sudo).\n    Execute: sudo %s", os.Args[0])
	}
	PrintOk("Running with sufficient privileges")
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
	PrintOk(fmt.Sprintf("%s mounted", caminho))
	return true
}

func resolveMediaMountInteractive() bool {
	if CheckMediaMount(MEDIA_DIR) {
		return true
	}
	PrintWarning(fmt.Sprintf(
		"%s não existe ou não é um mount separado. Ter um disco dedicado "+
			"é opcional — você pode seguir usando esse caminho como uma pasta comum no disco raiz.",
		MEDIA_DIR))
	return strings.ToLower(readLine("Continuar mesmo assim? [s/N]: ")) == "s"
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
		return fmt.Errorf("Não foi possível verificar o espaço em %s: %v", dataPath, err)
	}
	if dataLivre < MIN_DISK_SPACE_DATA_GB {
		return fmt.Errorf("Espaço insuficiente em %s: %.1fGB livres, mínimo exigido %dGB.",
			dataPath, dataLivre, MIN_DISK_SPACE_DATA_GB)
	}
	slog.Info("Espaco livre", "caminho", dataPath, "gb", fmt.Sprintf("%.1f", dataLivre))

	mediaPath := primeiroExistente(MEDIA_DIR)
	mediaLivre, err := espacoLivreGB(mediaPath)
	if err != nil {
		return fmt.Errorf("Não foi possível verificar o espaço em %s: %v", mediaPath, err)
	}
	slog.Info("Espaco livre", "caminho", mediaPath, "gb", fmt.Sprintf("%.1f", mediaLivre))
	if mediaLivre < MIN_DISK_SPACE_MEDIA_GB {
		PrintWarning(fmt.Sprintf("Pouco espaço livre em %s: %.1fGB (recomendado: %dGB+). Continuando mesmo assim.",
			mediaPath, mediaLivre, MIN_DISK_SPACE_MEDIA_GB))
	} else {
		PrintOk(fmt.Sprintf("Storage available (%.1fGB free on %s)", mediaLivre, mediaPath))
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
		PrintWarning("Não foi possível extrair portas dos compose files, usando lista padrão.")
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
		slog.Warn("Portas ocupadas", "portas", strings.Join(lista, ", "))
		PrintFail("Portas em uso detectadas:")
		for _, porta := range ocupadas {
			PrintInfo(fmt.Sprintf("porta %d (%s) já está em uso", porta, esperadas[porta]))
		}
		return errors.New("Libere as portas acima (pare o serviço que as está usando) ou ajuste " +
			"o docker-compose.yml antes de continuar.")
	}
	PrintOk("Required ports available")
	return nil
}
