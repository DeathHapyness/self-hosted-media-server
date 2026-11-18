package installer

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// O Loger (auditoria do servidor) roda direto no host, pelo systemd, e não
// como container: ele precisa enxergar os processos, sessões e arquivos do
// próprio host. O binário é extraído da imagem publicada no GHCR (sem rodar
// container nenhum), o que já entrega a arquitetura certa (amd64/arm64).

const (
	LOGER_IMAGE   = "ghcr.io/deathhapyness/loger:latest"
	LOGER_BIN     = "/usr/local/bin/loger"
	LOGER_UNIT    = "/etc/systemd/system/loger.service"
	LOGER_LOG_DIR = "/var/log/loger"
)

//go:embed loger/loger.service
var logerUnit string

func temSystemd() bool {
	return isDir("/run/systemd/system") && temBinario("systemctl")
}

func logerInstalado() bool {
	return existe(LOGER_BIN) || existe(LOGER_UNIT)
}

// perguntarLoger pergunta se o Loger deve ser instalado. Ele vem ativado:
// Enter sem nada mantém.
func perguntarLoger() bool {
	PrintSection(Tr("Auditoria do servidor (Loger)", "Server auditing (Loger)"))
	PrintInfo(Tr("O Loger vem ativado por padrão. Ele registra logins, comandos, uso de sudo, "+
		"alterações em arquivos sensíveis, chaves SSH e portas novas.",
		"Loger is enabled by default. It records logins, commands, sudo usage, "+
			"changes to sensitive files, SSH keys and new ports."))
	PrintInfo(fmt.Sprintf(Tr("Roda direto no host como serviço do systemd (%s) e grava os logs em %s. "+
		"Detalhes em docs/loger.md.",
		"It runs directly on the host as a systemd service (%s) and writes its logs to %s. "+
			"Details in docs/loger.md."), LOGER_BIN, LOGER_LOG_DIR))

	resposta := readLine(Tr("\nManter o Loger ativado? [S/n]: ", "\nKeep Loger enabled? [Y/n]: "))
	ativado := Confirmou(resposta, true)
	if ativado {
		PrintOk(Tr("Loger será instalado", "Loger will be installed"))
	} else {
		PrintInfo(Tr("Loger desativado: não será instalado", "Loger disabled: it will not be installed"))
	}
	Registrar("loger", "enabled", ativado)
	return ativado
}

// instalarLoger instala o binário e o serviço do systemd. Falhas aqui não
// interrompem a instalação: o Loger é opcional e os outros serviços já subiram.
func instalarLoger() {
	PrintSection(Tr("Instalando o Loger...", "Installing Loger..."))

	if !temSystemd() {
		PrintWarning(Tr("systemd não encontrado: o Loger não foi instalado. "+
			"Para outro init, deixe 'loger daemon' rodando como root (veja docs/loger.md).",
			"systemd not found: Loger was not installed. "+
				"With another init, keep 'loger daemon' running as root (see docs/loger.md)."))
		Registrar("loger", "result", "skipped", "reason", "no systemd")
		return
	}

	if DRY_RUN {
		if existe(LOGER_BIN) {
			PrintInfo(Tr("[dry-run] manteria o binário existente em ", "[dry-run] would keep the existing binary at ") + LOGER_BIN)
		} else {
			PrintInfo(fmt.Sprintf(Tr("[dry-run] extrairia /loger de %s para %s", "[dry-run] would extract /loger from %s to %s"), LOGER_IMAGE, LOGER_BIN))
		}
		if !existe(LOGER_UNIT) {
			PrintInfo(Tr("[dry-run] criaria ", "[dry-run] would create ") + LOGER_UNIT)
		}
		PrintInfo(Tr("[dry-run] rodaria: ", "[dry-run] would run: ") + "systemctl daemon-reload && systemctl enable --now loger")
		return
	}

	if existe(LOGER_BIN) {
		versao, _, _ := run(LOGER_BIN, "version")
		PrintOk(fmt.Sprintf(Tr("Binário já presente (%s), mantido", "Binary already present (%s), kept"), strings.TrimSpace(versao)))
	} else if err := extrairBinarioLoger(); err != nil {
		PrintWarning(Tr("Não foi possível instalar o binário do Loger: ", "Could not install the Loger binary: ") + err.Error())
		Registrar("loger", "result", "failed", "error", err.Error())
		return
	}

	// Não sobrescreve um loger.service existente: pode ter sido ajustado à mão.
	if existe(LOGER_UNIT) {
		PrintOk(LOGER_UNIT + Tr(" já existe, mantido", " already exists, kept"))
	} else if err := os.WriteFile(LOGER_UNIT, []byte(logerUnit), 0o644); err != nil {
		PrintWarning(Tr("Não foi possível criar ", "Could not create ") + LOGER_UNIT + ": " + err.Error())
		Registrar("loger", "result", "failed", "error", err.Error())
		return
	} else {
		PrintOk(LOGER_UNIT + Tr(" criado", " created"))
	}

	if _, stderr, err := run("systemctl", "daemon-reload"); err != nil {
		PrintWarning(Tr("systemctl daemon-reload falhou: ", "systemctl daemon-reload failed: ") + stderr)
	}
	if _, stderr, err := run("systemctl", "enable", "--now", "loger"); err != nil {
		PrintWarning(Tr("Não foi possível iniciar o Loger: ", "Could not start Loger: ") + stderr)
		PrintInfo(Tr("Veja o motivo com: ", "See why with: ") + "journalctl -u loger -n 50")
		Registrar("loger", "result", "failed", "error", stderr)
		return
	}
	PrintOk(Tr("Loger ativo (systemctl status loger). Consulte com: sudo loger show -n 50",
		"Loger active (systemctl status loger). Query it with: sudo loger show -n 50"))
	Registrar("loger", "result", "installed")
}

// extrairBinarioLoger baixa a imagem e copia /loger dela para LOGER_BIN.
// O container é só criado (nunca iniciado) e removido em seguida.
func extrairBinarioLoger() error {
	PrintInfo(Tr("Baixando ", "Downloading ") + LOGER_IMAGE + "...")
	if _, stderr, err := run(CONTAINER_ENGINE, "pull", LOGER_IMAGE); err != nil {
		return fmt.Errorf(Tr("falha ao baixar a imagem: %s", "failed to pull the image: %s"), stderr)
	}

	stdout, stderr, err := run(CONTAINER_ENGINE, "create", LOGER_IMAGE)
	if err != nil {
		return fmt.Errorf(Tr("falha ao preparar a imagem: %s", "failed to prepare the image: %s"), stderr)
	}
	id := strings.TrimSpace(stdout)
	defer run(CONTAINER_ENGINE, "rm", id)

	// Copia para um temporário na mesma pasta e renomeia: nunca deixa um
	// binário pela metade em LOGER_BIN.
	if err := os.MkdirAll(filepath.Dir(LOGER_BIN), 0o755); err != nil {
		return err
	}
	tmp := LOGER_BIN + ".novo"
	defer os.Remove(tmp)
	if _, stderr, err := run(CONTAINER_ENGINE, "cp", id+":/loger", tmp); err != nil {
		return fmt.Errorf(Tr("falha ao copiar o binário: %s", "failed to copy the binary: %s"), stderr)
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	versao, stderr, err := run(tmp, "version")
	if err != nil {
		return fmt.Errorf(Tr("o binário extraído não executa neste sistema: %s", "the extracted binary does not run on this system: %s"), stderr)
	}
	if err := os.Rename(tmp, LOGER_BIN); err != nil {
		return err
	}
	PrintOk(fmt.Sprintf(Tr("%s instalado em %s", "%s installed at %s"), strings.TrimSpace(versao), LOGER_BIN))
	return nil
}

// desinstalarLoger para e remove o serviço e o binário. Os logs ficam:
// são append-only (chattr +a) e servem de evidência.
func desinstalarLoger() {
	if !logerInstalado() {
		return
	}
	PrintSection(Tr("Loger (auditoria do servidor)", "Loger (server auditing)"))
	resposta := readLine(fmt.Sprintf(Tr("Também desinstalar o Loger? Os logs em %s são mantidos. [s/N]: ",
		"Also uninstall Loger? The logs in %s are kept. [y/N]: "), LOGER_LOG_DIR))
	if !Confirmou(resposta, false) {
		PrintInfo(Tr("Loger mantido.", "Loger kept."))
		Registrar("loger", "uninstall", "kept")
		return
	}

	if DRY_RUN {
		PrintInfo(Tr("[dry-run] rodaria: ", "[dry-run] would run: ") + "systemctl disable --now loger")
		PrintInfo(fmt.Sprintf(Tr("[dry-run] removeria %s e %s", "[dry-run] would remove %s and %s"), LOGER_UNIT, LOGER_BIN))
		return
	}

	if temSystemd() {
		if _, stderr, err := run("systemctl", "disable", "--now", "loger"); err != nil {
			slog.Warn("systemctl disable loger failed", "stderr", stderr)
		}
	}
	for _, alvo := range []string{LOGER_UNIT, LOGER_BIN} {
		if !existe(alvo) {
			continue
		}
		if err := os.Remove(alvo); err != nil {
			PrintWarning(alvo + Tr(" não pôde ser removido: ", " could not be removed: ") + err.Error())
		} else {
			PrintOk(alvo + Tr(" removido", " removed"))
		}
	}
	if temSystemd() {
		run("systemctl", "daemon-reload")
	}
	PrintInfo(fmt.Sprintf(Tr("Logs mantidos em %s (para apagar: sudo chattr -R -a %s && sudo rm -rf %s).",
		"Logs kept in %s (to delete them: sudo chattr -R -a %s && sudo rm -rf %s)."),
		LOGER_LOG_DIR, LOGER_LOG_DIR, LOGER_LOG_DIR))
	Registrar("loger", "uninstall", "removed")
}
