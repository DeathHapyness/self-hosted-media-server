package installer

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func CheckOS() error {
	if runtime.GOOS == "linux" {
		log.Printf("Sistema operacional: %s", runtime.GOOS)
		fmt.Println("Linux detectado")
		return nil
	}
	return fmt.Errorf("sistema operacional não suportado: %s. Este instalador requer Linux", runtime.GOOS)
}

func CheckPrivileges() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("este instalador precisa ser executado como root (sudo)\n    Execute: sudo %s", os.Args[0])
	}
	log.Printf("Rodando como root (uid %d)", os.Geteuid())
	fmt.Println("Rodando com privilégios suficientes")
	return nil
}

// CONFERE SE A PASTA DE MEDIA EXISTE
func CheckMediaMount(caminho string) bool {
	info, err := os.Stat(caminho)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Printf("%s não existe. Use a opção 'Gerenciar disco de mídia' no \n"+
			"menu principal para criar/montar, ou monte manualmente antes de continuar.\n", caminho)
		fmt.Println("O uso de um disco dedicado é opcional")
		return false
	}
	if err != nil {
		fmt.Printf("erro ao verificar %s: %v\n", caminho, err)
		return false
	}
	if !info.IsDir() {
		fmt.Printf("%s existe, mas não é uma pasta\n", caminho)
		return false
	}
	log.Printf("Pasta de mídia encontrada: %s", caminho)
	return true
}

func check_disk_space(caminho string) error {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		fmt.Printf("erro ao verificar %s: %v\n", "/", err)
		return err
	}
	// verifica se tem espaço no disco
	livre := stat.Bavail * uint64(stat.Bsize)
	livreGB := float64(livre) / (1024 * 1024 * 1024)
	fmt.Printf("Espaço livre em /: %.1fGB\n", livreGB)
	if livreGB < 20 {
		fmt.Print("Pouco espaço livre.\n" +
			"Recomendamos 30/40GB.")
	}
	return nil
}

func checkPorts(portas map[int]string) error {
	var ocupadas []string

	for porta, servico := range portas {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", porta), 500*time.Millisecond)
		if err == nil {
			conn.Close()
			ocupadas = append(ocupadas, fmt.Sprintf("%d (%s)", porta, servico))
		}
	}
	if len(ocupadas) > 0 {
		return fmt.Errorf("Portas em uso: %s", strings.Join(ocupadas, ","))
	}
	fmt.Println("Portas necessarias disponiveis")
	return nil
}
