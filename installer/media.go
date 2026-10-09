package installer

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const virtualDiskPath = "/var/media-disk.img"

type BlockDevice struct {
	Name       string // NAME: ex. "sda1"
	Size       string // SIZE: ex. "500G"
	Type       string // TYPE: "disk", "part", "loop"...
	MountPoint string // MOUNTPOINT: vazio se não estiver montado
	Parent     string // PKNAME: disco pai da partição; vazio se for um disco
}

func (d BlockDevice) Path() string {
	return "/dev/" + d.Name
}

func splitPairs(line string) []string {
	var pares []string
	var atual strings.Builder
	dentroDeAspas := false

	for _, r := range line {
		switch {
		case r == '"':
			dentroDeAspas = !dentroDeAspas
			atual.WriteRune(r)
		case r == ' ' && !dentroDeAspas:
			if atual.Len() > 0 {
				pares = append(pares, atual.String())
				atual.Reset()
			}
		default:
			atual.WriteRune(r)
		}
	}
	if atual.Len() > 0 {
		pares = append(pares, atual.String())
	}
	return pares
}

func parseLsblk(output string) []BlockDevice {
	var devices []BlockDevice
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var device BlockDevice
		for _, par := range splitPairs(line) {
			key, value, found := strings.Cut(par, "=")
			if !found {
				continue
			}
			value = strings.Trim(value, `"`)

			switch key {
			case "NAME":
				device.Name = value
			case "SIZE":
				device.Size = value
			case "TYPE":
				device.Type = value
			case "MOUNTPOINT":
				device.MountPoint = value
			case "PKNAME":
				device.Parent = value
			}
		}
		if device.Name != "" {
			devices = append(devices, device)
		}
	}
	return devices
}

func ListUnmountedBlockDevices() ([]BlockDevice, error) {
	result, _, err := run("lsblk", "-Pno", "NAME,SIZE,TYPE,MOUNTPOINT,PKNAME")
	if err != nil {
		return nil, err
	}
	slog.Info("lsblk", "saida", result)
	all := parseLsblk(result)

	hasChildren := map[string]bool{}
	for _, d := range all {
		if d.Parent != "" {
			hasChildren[d.Parent] = true
		}
	}

	var unmounted []BlockDevice
	for _, d := range all {
		if d.Type != "disk" && d.Type != "part" {
			continue
		}
		if d.MountPoint != "" {
			continue
		}
		if hasChildren[d.Name] {
			continue
		}
		unmounted = append(unmounted, d)
	}
	return unmounted, nil
}

func MountRealDeviceMenu() {
	PrintSection("Discos/partições disponíveis (sem uso no momento):")
	devices, err := ListUnmountedBlockDevices()
	if err != nil || len(devices) == 0 {
		PrintWarning("Nenhum disco/partição livre foi encontrado (ou 'lsblk' não está disponível).")
		return
	}

	for i, d := range devices {
		PrintInfo(fmt.Sprintf("%d) %s (%s)", i+1, d.Path(), d.Size))
	}

	escolha := readLine("\nDigite o número do dispositivo (ou vazio para cancelar): ")
	if escolha == "" {
		PrintInfo("Cancelado.")
		return
	}

	n, err := strconv.Atoi(escolha)
	if err != nil || n < 1 || n > len(devices) {
		PrintFail("Opção inválida.")
		return
	}
	device := devices[n-1]
	devicePath := device.Path()

	aviso := fmt.Sprintf(
		"Você está prestes a FORMATAR %s (%s) como ext4.\n"+
			"    TODOS OS DADOS atualmente nesse dispositivo serão APAGADOS PERMANENTEMENTE.\n"+
			"    Confirme que %s é realmente o disco certo antes de continuar.",
		devicePath, device.Size, devicePath)
	if !confirmDangerous(aviso, devicePath) {
		PrintInfo("Operação cancelada — nada foi alterado.")
		registrar("formatar_disco", "dispositivo", devicePath, "resultado", "cancelado")
		return
	}

	if DRY_RUN {
		PrintOk(fmt.Sprintf("[dry-run] formataria %s como ext4 e montaria em %s", devicePath, MEDIA_DIR))
		registrar("formatar_disco", "dispositivo", devicePath, "resultado", "dry-run")
		return
	}

	if err := os.MkdirAll(MEDIA_DIR, 0o755); err != nil {
		PrintFail(fmt.Sprintf("Não foi possível criar %s: %v", MEDIA_DIR, err))
		return
	}

	PrintSection(fmt.Sprintf("Formatando %s...", devicePath))
	if _, stderr, err := run("mkfs.ext4", "-F", devicePath); err != nil {
		PrintFail("Falha ao formatar:")
		PrintInfo(stderr)
		registrar("formatar_disco", "dispositivo", devicePath, "resultado", "falha ao formatar", "erro", stderr)
		return
	}
	PrintOk("Formatado com sucesso")

	if _, stderr, err := run("mount", devicePath, MEDIA_DIR); err != nil {
		PrintFail("Falha ao montar:")
		PrintInfo(stderr)
		registrar("formatar_disco", "dispositivo", devicePath, "resultado", "falha ao montar", "erro", stderr)
		return
	}
	PrintOk(fmt.Sprintf("%s montado em %s", devicePath, MEDIA_DIR))
	registrar("formatar_disco", "dispositivo", devicePath, "resultado", "montado em "+MEDIA_DIR)

	adicionarFstab(fmt.Sprintf("%s %s ext4 defaults 0 2", devicePath, MEDIA_DIR))
}

func CreateVirtualDiskMenu() {
	PrintSection(fmt.Sprintf("Criar disco virtual (arquivo de imagem) para %s", MEDIA_DIR))
	PrintInfo("Isso ocupa espaço no disco atual — só recomendado para testes ou quando")
	PrintInfo("não há disco físico disponível. Para produção, prefira um disco real.")

	tamanho := readLine("\nTamanho em GB (padrão 20): ")
	if tamanho == "" {
		tamanho = "20"
	}
	tamanhoGB, err := strconv.Atoi(tamanho)
	if err != nil || tamanhoGB <= 0 {
		PrintFail("Tamanho inválido.")
		return
	}

	aviso := fmt.Sprintf(
		"Será criado um arquivo de %dGB em %s e montado em %s.\n"+
			"    Se já existir um disco virtual anterior nesse caminho, ele será sobrescrito.",
		tamanhoGB, virtualDiskPath, MEDIA_DIR)
	if !confirmDangerous(aviso, "CONFIRMO") {
		PrintInfo("Operação cancelada — nada foi alterado.")
		registrar("criar_disco_virtual", "tamanho_gb", tamanhoGB, "resultado", "cancelado")
		return
	}

	if DRY_RUN {
		PrintOk(fmt.Sprintf("[dry-run] criaria %s (%dGB) e montaria em %s", virtualDiskPath, tamanhoGB, MEDIA_DIR))
		registrar("criar_disco_virtual", "tamanho_gb", tamanhoGB, "resultado", "dry-run")
		return
	}

	if err := os.MkdirAll(MEDIA_DIR, 0o755); err != nil {
		PrintFail(fmt.Sprintf("Não foi possível criar %s: %v", MEDIA_DIR, err))
		return
	}

	PrintSection("Criando arquivo de disco virtual...")
	if _, _, err := run("fallocate", "-l", fmt.Sprintf("%dG", tamanhoGB), virtualDiskPath); err != nil {
		_, stderr, err := run("dd", "if=/dev/zero", "of="+virtualDiskPath, "bs=1M",
			fmt.Sprintf("count=%d", tamanhoGB*1024))
		if err != nil {
			PrintFail("Falha ao criar o arquivo de disco virtual:")
			PrintInfo(stderr)
			registrar("criar_disco_virtual", "tamanho_gb", tamanhoGB, "resultado", "falha ao criar arquivo", "erro", stderr)
			return
		}
	}
	PrintOk(fmt.Sprintf("Arquivo criado (%dGB)", tamanhoGB))

	if _, stderr, err := run("mkfs.ext4", "-F", virtualDiskPath); err != nil {
		PrintFail("Falha ao formatar:")
		PrintInfo(stderr)
		registrar("criar_disco_virtual", "tamanho_gb", tamanhoGB, "resultado", "falha ao formatar", "erro", stderr)
		return
	}
	PrintOk("Formatado como ext4")

	if _, stderr, err := run("mount", "-o", "loop", virtualDiskPath, MEDIA_DIR); err != nil {
		PrintFail("Falha ao montar:")
		PrintInfo(stderr)
		registrar("criar_disco_virtual", "tamanho_gb", tamanhoGB, "resultado", "falha ao montar", "erro", stderr)
		return
	}
	PrintOk(fmt.Sprintf("Montado em %s", MEDIA_DIR))
	registrar("criar_disco_virtual", "tamanho_gb", tamanhoGB, "resultado", "montado em "+MEDIA_DIR)

	adicionarFstab(fmt.Sprintf("%s %s ext4 loop 0 0", virtualDiskPath, MEDIA_DIR))
}

func adicionarFstab(linha string) {
	const fstab = "/etc/fstab"

	conteudo, err := os.ReadFile(fstab)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		PrintFail(fmt.Sprintf("Não foi possível ler %s: %v", fstab, err))
		return
	}
	if strings.Contains(string(conteudo), linha) {
		PrintOk("/etc/fstab já contém essa entrada")
		return
	}

	resposta := strings.ToLower(readLine("\nAdicionar entrada em /etc/fstab para sobreviver a um reboot? [s/N]: "))
	if resposta != "s" {
		PrintInfo("Pulado — o mount não sobrevive a um reboot até você adicionar manualmente.")
		registrar("fstab", "linha", linha, "resultado", "pulado")
		return
	}

	slog.Warn("Adicionando entrada em /etc/fstab", "linha", linha)
	f, err := os.OpenFile(fstab, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		PrintFail(fmt.Sprintf("Não foi possível abrir %s: %v", fstab, err))
		return
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "\n%s\n", linha); err != nil {
		PrintFail(fmt.Sprintf("Não foi possível gravar em %s: %v", fstab, err))
		return
	}
	PrintOk("/etc/fstab atualizado")
	registrar("fstab", "linha", linha, "resultado", "adicionada")
}

func MediaMountMenu() {
	for {
		PrintHeader("GESTÃO DO DISCO DE MÍDIA")
		status := ColorText("NÃO montado", RED)
		if CheckMediaMount(MEDIA_DIR) {
			status = ColorText("montado", GREEN)
		}
		fmt.Printf("Status atual de %s: %s\n\n", MEDIA_DIR, status)

		PrintInfo("1) Já montei manualmente — só verificar")
		PrintInfo("2) Usar um disco/partição existente (formata e monta — APAGA DADOS do disco escolhido)")
		PrintInfo("3) Criar um disco virtual (arquivo, menos arriscado, ocupa espaço do disco atual)")
		PrintInfo("4) Voltar ao menu principal")

		switch readLine("\nEscolha uma opção: ") {
		case "1":
			CheckMediaMount(MEDIA_DIR)
			pressEnterToContinue()
		case "2":
			MountRealDeviceMenu()
			pressEnterToContinue()
		case "3":
			CreateVirtualDiskMenu()
			pressEnterToContinue()
		case "4":
			return
		default:
			PrintFail("Opção inválida.")
		}
	}
}
