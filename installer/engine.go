package installer

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func temBinario(nome string) bool {
	_, err := exec.LookPath(nome)
	return err == nil
}

func chooseEngine() string {
	dockerOk := temBinario("docker")
	podmanOk := temBinario("podman")

	if dockerOk && podmanOk {
		PrintSection(Tr("Docker e Podman detectados neste sistema.", "Docker and Podman detected on this system."))
		PrintInfo(Tr("1) Docker (padrão)", "1) Docker (default)"))
		PrintInfo("2) Podman")
		if readLine(Tr("Qual usar? [1]: ", "Which one to use? [1]: ")) == "2" {
			return "podman"
		}
		return "docker"
	}

	if podmanOk && !dockerOk {
		return "podman"
	}
	return "docker"
}

func checkEngineBinary() error {
	binary := CONTAINER_ENGINE
	if !temBinario(binary) {
		if binary == "podman" {
			return errors.New(Tr("Podman não encontrado. Instale com o gerenciador de pacotes da sua "+
				"distro (ex: apt install podman) ou veja: https://podman.io/docs/installation",
				"Podman not found. Install it with your distro's package manager "+
					"(e.g. apt install podman) or see: https://podman.io/docs/installation"))
		}
		return errors.New(Tr("Docker não encontrado.\n"+
			"    Instale com: curl -fsSL https://get.docker.com | sh\n"+
			"    Ou veja: https://docs.docker.com/engine/install/\n"+
			"    Fonte do script: https://github.com/docker/docker-install",
			"Docker not found.\n"+
				"    Install it with: curl -fsSL https://get.docker.com | sh\n"+
				"    Or see: https://docs.docker.com/engine/install/\n"+
				"    Script source: https://github.com/docker/docker-install"))
	}
	stdout, _, err := run(binary, "--version")
	if err != nil {
		return fmt.Errorf(Tr("%s está instalado, mas '%s --version' falhou.", "%s is installed, but '%s --version' failed."), binary, binary)
	}
	versao := strings.TrimSpace(stdout)
	slog.Info("Engine", "binary", binary, "version", versao)
	PrintOk(fmt.Sprintf(Tr("%s instalado (%s)", "%s installed (%s)"), capitalize(binary), versao))
	return nil
}

func checkComposePlugin() error {
	engine := CONTAINER_ENGINE
	stdout, _, err := run(engine, "compose", "version")
	if err == nil {
		Compose_Standalone = false
		PrintOk(fmt.Sprintf(Tr("%s Compose instalado (%s)", "%s Compose installed (%s)"), capitalize(engine), strings.TrimSpace(stdout)))
		return nil
	}

	if engine == "podman" && temBinario("podman-compose") {
		stdout, _, err := run("podman-compose", "--version")
		if err == nil {
			Compose_Standalone = true
			PrintOk(fmt.Sprintf(Tr("podman-compose instalado (%s)", "podman-compose installed (%s)"), strings.TrimSpace(stdout)))
			return nil
		}
	}

	return fmt.Errorf(Tr("Compose não encontrado para %s. Instale o plugin '%s compose' "+
		"ou, no caso do Podman, o pacote podman-compose.",
		"Compose not found for %s. Install the '%s compose' plugin "+
			"or, for Podman, the podman-compose package."), engine, engine)
}

func checkEngineDaemon() error {
	engine := CONTAINER_ENGINE
	if _, _, err := run(engine, "info"); err != nil {
		return fmt.Errorf(Tr("%s não está rodando ou o usuário atual não tem permissão. "+
			"Tente 'sudo systemctl start %s' ou rode este script com sudo.",
			"%s is not running or the current user lacks permission. "+
				"Try 'sudo systemctl start %s' or run this script with sudo."), capitalize(engine), engine)
	}
	PrintOk(fmt.Sprintf(Tr("Daemon do %s rodando", "%s daemon running"), capitalize(engine)))
	return nil
}

func composeBaseCmd() []string {
	if CONTAINER_ENGINE == "podman" && Compose_Standalone {
		return []string{"podman-compose"}
	}
	return []string{CONTAINER_ENGINE, "compose"}
}

func checkOptional(checkFn func() error, label string) bool {
	err := checkFn()
	if err == nil {
		return true
	}
	PrintFail(err.Error())
	resposta := readLine(fmt.Sprintf(Tr("\nContinuar mesmo assim, pulando a verificação de %s? [s/N]: ",
		"\nContinue anyway, skipping the %s check? [y/N]: "), label))
	if Confirmou(resposta, false) {
		slog.Warn("User chose to skip a failed check", "check", label, "error", err)
		PrintWarning(fmt.Sprintf(Tr("Continuando sem validar %s — por sua conta e risco.",
			"Continuing without validating %s — at your own risk."), label))
		return true
	}
	return false
}
