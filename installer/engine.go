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
		PrintSection("Docker e Podman detectados neste sistema.")
		PrintInfo("1) Docker (padrão)")
		PrintInfo("2) Podman")
		if readLine("Qual usar? [1]: ") == "2" {
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
			return errors.New("Podman não encontrado. Instale com o gerenciador de pacotes da sua " +
				"distro (ex: apt install podman) ou veja: https://podman.io/docs/installation")
		}
		return errors.New("Docker não encontrado.\n" +
			"    Instale com: curl -fsSL https://get.docker.com | sh\n" +
			"    Ou veja: https://docs.docker.com/engine/install/\n" +
			"    Fonte do script: https://github.com/docker/docker-install")
	}
	stdout, _, err := run(binary, "--version")
	if err != nil {
		return fmt.Errorf("%s está instalado, mas '%s --version' falhou.", binary, binary)
	}
	versao := strings.TrimSpace(stdout)
	slog.Info("Engine", "binario", binary, "versao", versao)
	PrintOk(fmt.Sprintf("%s installed (%s)", capitalize(binary), versao))
	return nil
}

func checkComposePlugin() error {
	engine := CONTAINER_ENGINE
	stdout, _, err := run(engine, "compose", "version")
	if err == nil {
		Compose_Standalone = false
		PrintOk(fmt.Sprintf("%s Compose installed (%s)", capitalize(engine), strings.TrimSpace(stdout)))
		return nil
	}

	if engine == "podman" && temBinario("podman-compose") {
		stdout, _, err := run("podman-compose", "--version")
		if err == nil {
			Compose_Standalone = true
			PrintOk(fmt.Sprintf("podman-compose installed (%s)", strings.TrimSpace(stdout)))
			return nil
		}
	}

	return fmt.Errorf("Compose não encontrado para %s. Instale o plugin '%s compose' "+
		"ou, no caso do Podman, o pacote podman-compose.", engine, engine)
}

func checkEngineDaemon() error {
	engine := CONTAINER_ENGINE
	if _, _, err := run(engine, "info"); err != nil {
		return fmt.Errorf("%s não está rodando ou o usuário atual não tem permissão. "+
			"Tente 'sudo systemctl start %s' ou rode este script com sudo.", capitalize(engine), engine)
	}
	PrintOk(fmt.Sprintf("%s daemon running", capitalize(engine)))
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
	resposta := strings.ToLower(readLine(fmt.Sprintf("\nContinuar mesmo assim, pulando a verificação de %s? [s/N]: ", label)))
	if resposta == "s" {
		slog.Warn("Usuario optou por pular a verificacao apos falha", "verificacao", label, "erro", err)
		PrintWarning(fmt.Sprintf("Continuando sem validar %s — por sua conta e risco.", label))
		return true
	}
	return false
}
