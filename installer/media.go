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
	Name       string
	Size       string
	Type       string
	MountPoint string
	Parent     string
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
	slog.Info("lsblk", "output", result)
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
	PrintSection(Tr("Discos/partições disponíveis (sem uso no momento):", "Available disks/partitions (currently unused):"))
	devices, err := ListUnmountedBlockDevices()
	if err != nil || len(devices) == 0 {
		PrintWarning(Tr("Nenhum disco/partição livre foi encontrado (ou 'lsblk' não está disponível).",
			"No free disk/partition found (or 'lsblk' is not available)."))
		return
	}

	for i, d := range devices {
		PrintInfo(fmt.Sprintf("%d) %s (%s)", i+1, d.Path(), d.Size))
	}

	escolha := readLine(Tr("\nDigite o número do dispositivo (ou vazio para cancelar): ", "\nType the device number (or empty to cancel): "))
	if escolha == "" {
		PrintInfo(Tr("Cancelado.", "Cancelled."))
		return
	}

	n, err := strconv.Atoi(escolha)
	if err != nil || n < 1 || n > len(devices) {
		PrintFail(Tr("Opção inválida.", "Invalid option."))
		return
	}
	device := devices[n-1]
	devicePath := device.Path()

	aviso := fmt.Sprintf(Tr(
		"Você está prestes a FORMATAR %s (%s) como ext4.\n"+
			"    TODOS OS DADOS atualmente nesse dispositivo serão APAGADOS PERMANENTEMENTE.\n"+
			"    Confirme que %s é realmente o disco certo antes de continuar.",
		"You are about to FORMAT %s (%s) as ext4.\n"+
			"    ALL DATA currently on this device will be PERMANENTLY ERASED.\n"+
			"    Make sure %s is really the right disk before continuing."),
		devicePath, device.Size, devicePath)
	if !confirmDangerous(aviso, devicePath) {
		PrintInfo(Tr("Operação cancelada — nada foi alterado.", "Operation cancelled — nothing was changed."))
		Registrar("format_disk", "device", devicePath, "result", "cancelled")
		return
	}

	if DRY_RUN {
		PrintOk(fmt.Sprintf(Tr("[dry-run] formataria %s como ext4 e montaria em %s", "[dry-run] would format %s as ext4 and mount it on %s"), devicePath, MEDIA_DIR))
		Registrar("format_disk", "device", devicePath, "result", "dry-run")
		return
	}

	if err := os.MkdirAll(MEDIA_DIR, 0o755); err != nil {
		PrintFail(fmt.Sprintf(Tr("Não foi possível criar %s: %v", "Could not create %s: %v"), MEDIA_DIR, err))
		return
	}

	PrintSection(fmt.Sprintf(Tr("Formatando %s...", "Formatting %s..."), devicePath))
	if _, stderr, err := run("mkfs.ext4", "-F", devicePath); err != nil {
		PrintFail(Tr("Falha ao formatar:", "Formatting failed:"))
		PrintInfo(stderr)
		Registrar("format_disk", "device", devicePath, "result", "format failed", "error", stderr)
		return
	}
	PrintOk(Tr("Formatado com sucesso", "Formatted successfully"))

	if _, stderr, err := run("mount", devicePath, MEDIA_DIR); err != nil {
		PrintFail(Tr("Falha ao montar:", "Mount failed:"))
		PrintInfo(stderr)
		Registrar("format_disk", "device", devicePath, "result", "mount failed", "error", stderr)
		return
	}
	PrintOk(fmt.Sprintf(Tr("%s montado em %s", "%s mounted on %s"), devicePath, MEDIA_DIR))
	Registrar("format_disk", "device", devicePath, "result", "mounted on "+MEDIA_DIR)

	adicionarFstab(fmt.Sprintf("%s %s ext4 defaults 0 2", devicePath, MEDIA_DIR))
}

func CreateVirtualDiskMenu() {
	PrintSection(fmt.Sprintf(Tr("Criar disco virtual (arquivo de imagem) para %s", "Create a virtual disk (image file) for %s"), MEDIA_DIR))
	PrintInfo(Tr("Isso ocupa espaço no disco atual — só recomendado para testes ou quando",
		"This uses space on the current disk — only recommended for testing or when"))
	PrintInfo(Tr("não há disco físico disponível. Para produção, prefira um disco real.",
		"no physical disk is available. For production, prefer a real disk."))

	tamanho := readLine(Tr("\nTamanho em GB (padrão 20): ", "\nSize in GB (default 20): "))
	if tamanho == "" {
		tamanho = "20"
	}
	tamanhoGB, err := strconv.Atoi(tamanho)
	if err != nil || tamanhoGB <= 0 {
		PrintFail(Tr("Tamanho inválido.", "Invalid size."))
		return
	}

	aviso := fmt.Sprintf(Tr(
		"Será criado um arquivo de %dGB em %s e montado em %s.\n"+
			"    Se já existir um disco virtual anterior nesse caminho, ele será sobrescrito.",
		"A %dGB file will be created at %s and mounted on %s.\n"+
			"    If a previous virtual disk already exists at that path, it will be overwritten."),
		tamanhoGB, virtualDiskPath, MEDIA_DIR)
	if !confirmDangerous(aviso, PalavraConfirmacao()) {
		PrintInfo(Tr("Operação cancelada — nada foi alterado.", "Operation cancelled — nothing was changed."))
		Registrar("create_virtual_disk", "size_gb", tamanhoGB, "result", "cancelled")
		return
	}

	if DRY_RUN {
		PrintOk(fmt.Sprintf(Tr("[dry-run] criaria %s (%dGB) e montaria em %s", "[dry-run] would create %s (%dGB) and mount it on %s"), virtualDiskPath, tamanhoGB, MEDIA_DIR))
		Registrar("create_virtual_disk", "size_gb", tamanhoGB, "result", "dry-run")
		return
	}

	if err := os.MkdirAll(MEDIA_DIR, 0o755); err != nil {
		PrintFail(fmt.Sprintf(Tr("Não foi possível criar %s: %v", "Could not create %s: %v"), MEDIA_DIR, err))
		return
	}

	PrintSection(Tr("Criando arquivo de disco virtual...", "Creating the virtual disk file..."))
	if _, _, err := run("fallocate", "-l", fmt.Sprintf("%dG", tamanhoGB), virtualDiskPath); err != nil {
		_, stderr, err := run("dd", "if=/dev/zero", "of="+virtualDiskPath, "bs=1M",
			fmt.Sprintf("count=%d", tamanhoGB*1024))
		if err != nil {
			PrintFail(Tr("Falha ao criar o arquivo de disco virtual:", "Failed to create the virtual disk file:"))
			PrintInfo(stderr)
			Registrar("create_virtual_disk", "size_gb", tamanhoGB, "result", "file creation failed", "error", stderr)
			return
		}
	}
	PrintOk(fmt.Sprintf(Tr("Arquivo criado (%dGB)", "File created (%dGB)"), tamanhoGB))

	if _, stderr, err := run("mkfs.ext4", "-F", virtualDiskPath); err != nil {
		PrintFail(Tr("Falha ao formatar:", "Formatting failed:"))
		PrintInfo(stderr)
		Registrar("create_virtual_disk", "size_gb", tamanhoGB, "result", "format failed", "error", stderr)
		return
	}
	PrintOk(Tr("Formatado como ext4", "Formatted as ext4"))

	if _, stderr, err := run("mount", "-o", "loop", virtualDiskPath, MEDIA_DIR); err != nil {
		PrintFail(Tr("Falha ao montar:", "Mount failed:"))
		PrintInfo(stderr)
		Registrar("create_virtual_disk", "size_gb", tamanhoGB, "result", "mount failed", "error", stderr)
		return
	}
	PrintOk(fmt.Sprintf(Tr("Montado em %s", "Mounted on %s"), MEDIA_DIR))
	Registrar("create_virtual_disk", "size_gb", tamanhoGB, "result", "mounted on "+MEDIA_DIR)

	adicionarFstab(fmt.Sprintf("%s %s ext4 loop 0 0", virtualDiskPath, MEDIA_DIR))
}

func adicionarFstab(linha string) {
	const fstab = "/etc/fstab"

	conteudo, err := os.ReadFile(fstab)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		PrintFail(fmt.Sprintf(Tr("Não foi possível ler %s: %v", "Could not read %s: %v"), fstab, err))
		return
	}
	if strings.Contains(string(conteudo), linha) {
		PrintOk(Tr("/etc/fstab já contém essa entrada", "/etc/fstab already has this entry"))
		return
	}

	resposta := readLine(Tr("\nAdicionar entrada em /etc/fstab para sobreviver a um reboot? [s/N]: ",
		"\nAdd an /etc/fstab entry so it survives a reboot? [y/N]: "))
	if !Confirmou(resposta, false) {
		PrintInfo(Tr("Pulado — o mount não sobrevive a um reboot até você adicionar manualmente.",
			"Skipped — the mount will not survive a reboot until you add it manually."))
		Registrar("fstab", "line", linha, "result", "skipped")
		return
	}

	slog.Warn("Adding /etc/fstab entry", "line", linha)
	f, err := os.OpenFile(fstab, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		PrintFail(fmt.Sprintf(Tr("Não foi possível abrir %s: %v", "Could not open %s: %v"), fstab, err))
		return
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "\n%s\n", linha); err != nil {
		PrintFail(fmt.Sprintf(Tr("Não foi possível gravar em %s: %v", "Could not write to %s: %v"), fstab, err))
		return
	}
	PrintOk(Tr("/etc/fstab atualizado", "/etc/fstab updated"))
	Registrar("fstab", "line", linha, "result", "added")
}

func MediaMountMenu() {
	for {
		PrintHeader(Tr("GESTÃO DO DISCO DE MÍDIA", "MEDIA DISK MANAGEMENT"))
		status := ColorText(Tr("NÃO montado", "NOT mounted"), RED)
		if CheckMediaMount(MEDIA_DIR) {
			status = ColorText(Tr("montado", "mounted"), GREEN)
		}
		PrintText(fmt.Sprintf(Tr("Status atual de %s: %s\n\n", "Current status of %s: %s\n\n"), MEDIA_DIR, status))

		PrintInfo(Tr("1) Já montei manualmente — só verificar", "1) I already mounted it manually — just check"))
		PrintInfo(Tr("2) Usar um disco/partição existente (formata e monta — APAGA DADOS do disco escolhido)",
			"2) Use an existing disk/partition (formats and mounts — ERASES DATA on the chosen disk)"))
		PrintInfo(Tr("3) Criar um disco virtual (arquivo, menos arriscado, ocupa espaço do disco atual)",
			"3) Create a virtual disk (file, less risky, uses space on the current disk)"))
		PrintInfo(Tr("4) Voltar ao menu principal", "4) Back to the main menu"))

		switch readLine(Tr("\nEscolha uma opção: ", "\nChoose an option: ")) {
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
			PrintFail(Tr("Opção inválida.", "Invalid option."))
		}
	}
}
