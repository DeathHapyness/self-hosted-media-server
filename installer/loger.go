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
	PrintSection("Auditoria do servidor (Loger)")
	PrintInfo("O Loger vem ativado por padrão. Ele registra logins, comandos, uso de sudo, " +
		"alterações em arquivos sensíveis, chaves SSH e portas novas.")
	PrintInfo(fmt.Sprintf("Roda direto no host como serviço do systemd (%s) e grava os logs em %s. "+
		"Detalhes em docs/loger.md.", LOGER_BIN, LOGER_LOG_DIR))

	resposta := strings.ToLower(readLine("\nManter o Loger ativado? [S/n]: "))
	ativado := resposta != "n" && resposta != "nao" && resposta != "não"
	if ativado {
		PrintOk("Loger será instalado")
	} else {
		PrintInfo("Loger desativado: não será instalado")
	}
	Registrar("loger", "ativado", ativado)
	return ativado
}

// instalarLoger instala o binário e o serviço do systemd. Falhas aqui não
// interrompem a instalação: o Loger é opcional e os outros serviços já subiram.
func instalarLoger() {
	PrintSection("Instalando o Loger...")

	if !temSystemd() {
		PrintWarning("systemd não encontrado: o Loger não foi instalado. " +
			"Para outro init, deixe 'loger daemon' rodando como root (veja docs/loger.md).")
		Registrar("loger", "resultado", "pulado", "motivo", "sem systemd")
		return
	}

	if DRY_RUN {
		if existe(LOGER_BIN) {
			PrintInfo("[dry-run] manteria o binário existente em " + LOGER_BIN)
		} else {
			PrintInfo(fmt.Sprintf("[dry-run] extrairia /loger de %s para %s", LOGER_IMAGE, LOGER_BIN))
		}
		if !existe(LOGER_UNIT) {
			PrintInfo("[dry-run] criaria " + LOGER_UNIT)
		}
		PrintInfo("[dry-run] rodaria: systemctl daemon-reload && systemctl enable --now loger")
		return
	}

	if existe(LOGER_BIN) {
		versao, _, _ := run(LOGER_BIN, "version")
		PrintOk(fmt.Sprintf("Binário já presente (%s), mantido", strings.TrimSpace(versao)))
	} else if err := extrairBinarioLoger(); err != nil {
		PrintWarning("Não foi possível instalar o binário do Loger: " + err.Error())
		Registrar("loger", "resultado", "falhou", "erro", err.Error())
		return
	}

	// Não sobrescreve um loger.service existente: pode ter sido ajustado à mão.
	if existe(LOGER_UNIT) {
		PrintOk(LOGER_UNIT + " já existe, mantido")
	} else if err := os.WriteFile(LOGER_UNIT, []byte(logerUnit), 0o644); err != nil {
		PrintWarning("Não foi possível criar " + LOGER_UNIT + ": " + err.Error())
		Registrar("loger", "resultado", "falhou", "erro", err.Error())
		return
	} else {
		PrintOk(LOGER_UNIT + " criado")
	}

	if _, stderr, err := run("systemctl", "daemon-reload"); err != nil {
		PrintWarning("systemctl daemon-reload falhou: " + stderr)
	}
	if _, stderr, err := run("systemctl", "enable", "--now", "loger"); err != nil {
		PrintWarning("Não foi possível iniciar o Loger: " + stderr)
		PrintInfo("Veja o motivo com: journalctl -u loger -n 50")
		Registrar("loger", "resultado", "falhou", "erro", stderr)
		return
	}
	PrintOk("Loger ativo (systemctl status loger). Consulte com: sudo loger show -n 50")
	Registrar("loger", "resultado", "instalado")
}

// extrairBinarioLoger baixa a imagem e copia /loger dela para LOGER_BIN.
// O container é só criado (nunca iniciado) e removido em seguida.
func extrairBinarioLoger() error {
	PrintInfo("Baixando " + LOGER_IMAGE + "...")
	if _, stderr, err := run(CONTAINER_ENGINE, "pull", LOGER_IMAGE); err != nil {
		return fmt.Errorf("falha ao baixar a imagem: %s", stderr)
	}

	stdout, stderr, err := run(CONTAINER_ENGINE, "create", LOGER_IMAGE)
	if err != nil {
		return fmt.Errorf("falha ao preparar a imagem: %s", stderr)
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
		return fmt.Errorf("falha ao copiar o binário: %s", stderr)
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	versao, stderr, err := run(tmp, "version")
	if err != nil {
		return fmt.Errorf("o binário extraído não executa neste sistema: %s", stderr)
	}
	if err := os.Rename(tmp, LOGER_BIN); err != nil {
		return err
	}
	PrintOk(fmt.Sprintf("%s instalado em %s", strings.TrimSpace(versao), LOGER_BIN))
	return nil
}

// desinstalarLoger para e remove o serviço e o binário. Os logs ficam:
// são append-only (chattr +a) e servem de evidência.
func desinstalarLoger() {
	if !logerInstalado() {
		return
	}
	PrintSection("Loger (auditoria do servidor)")
	resposta := strings.ToLower(readLine("Também desinstalar o Loger? Os logs em " +
		LOGER_LOG_DIR + " são mantidos. [s/N]: "))
	if resposta != "s" {
		PrintInfo("Loger mantido.")
		Registrar("loger", "desinstalacao", "mantido")
		return
	}

	if DRY_RUN {
		PrintInfo("[dry-run] rodaria: systemctl disable --now loger")
		PrintInfo(fmt.Sprintf("[dry-run] removeria %s e %s", LOGER_UNIT, LOGER_BIN))
		return
	}

	if temSystemd() {
		if _, stderr, err := run("systemctl", "disable", "--now", "loger"); err != nil {
			slog.Warn("systemctl disable loger falhou", "stderr", stderr)
		}
	}
	for _, alvo := range []string{LOGER_UNIT, LOGER_BIN} {
		if !existe(alvo) {
			continue
		}
		if err := os.Remove(alvo); err != nil {
			PrintWarning(alvo + " não pôde ser removido: " + err.Error())
		} else {
			PrintOk(alvo + " removido")
		}
	}
	if temSystemd() {
		run("systemctl", "daemon-reload")
	}
	PrintInfo(fmt.Sprintf("Logs mantidos em %s (para apagar: sudo chattr -R -a %s && sudo rm -rf %s).",
		LOGER_LOG_DIR, LOGER_LOG_DIR, LOGER_LOG_DIR))
	Registrar("loger", "desinstalacao", "removido")
}
