package installer

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// O DockerBuddy é uma colinha de comandos Docker para o terminal: um script
// bash com um menu numerado que mostra os comandos mais usados (containers,
// imagens, volumes, redes, Compose e filtros do `docker ps`), cada um com uma
// explicação curta. Não é container nem serviço: não abre porta e não fica
// rodando. Ele só imprime texto quando você chama `dockerbuddy`.
//
// O instalador baixa o código do repositório (só o script, os .txt de data/,
// o README e a LICENSE) para DOCKERBUDDY_DIR e cria um link em DOCKERBUDDY_BIN.

const (
	DOCKERBUDDY_URL = "https://github.com/DeathHapyness/DockerBuddy/archive/refs/heads/main.tar.gz"
	DOCKERBUDDY_DIR = "/opt/dockerbuddy"
	DOCKERBUDDY_BIN = "/usr/local/bin/dockerbuddy"

	// O repositório tem poucos KB de texto; o limite só evita baixar algo
	// enorme se a URL passar a apontar para outra coisa.
	dockerbuddyMaxBytes = 10 << 20
)

func dockerbuddyScript() string {
	return filepath.Join(DOCKERBUDDY_DIR, "dockerbuddy.sh")
}

func dockerbuddyInstalado() bool {
	_, err := os.Lstat(DOCKERBUDDY_BIN)
	return err == nil || existe(DOCKERBUDDY_DIR)
}

// perguntarDockerBuddy explica o que é o DockerBuddy e pergunta se deve ser
// instalado. Vem marcado: Enter sem nada instala.
func perguntarDockerBuddy() bool {
	PrintSection(Tr("Colinha de comandos Docker (DockerBuddy)", "Docker command cheat sheet (DockerBuddy)"))
	PrintInfo(Tr("O DockerBuddy é uma colinha de comandos Docker no terminal. Você digita "+
		"'dockerbuddy', escolhe um número e ele mostra os comandos mais usados, cada um "+
		"com uma explicação curta: containers, imagens, volumes, redes, Docker Compose "+
		"e filtros do 'docker ps'.",
		"DockerBuddy is a Docker command cheat sheet for the terminal. You type "+
			"'dockerbuddy', pick a number and it shows the most used commands, each "+
			"with a short explanation: containers, images, volumes, networks, Docker Compose "+
			"and 'docker ps' filters."))
	PrintInfo(Tr("Serve para o dia a dia do servidor: ver logs, reiniciar um serviço, "+
		"limpar imagens velhas, sem precisar procurar o comando na internet.",
		"Handy for day-to-day server work: reading logs, restarting a service, "+
			"cleaning up old images, without searching the web for the command."))
	PrintInfo(fmt.Sprintf(Tr("Não é container nem serviço: não abre porta e não fica rodando. "+
		"São só um script e alguns arquivos de texto em %s, com o comando %s. "+
		"Detalhes em docs/dockerbuddy.md.",
		"It is not a container or a service: it opens no port and does not keep running. "+
			"It is just a script and a few text files in %s, with the %s command. "+
			"Details in docs/dockerbuddy.md."), DOCKERBUDDY_DIR, DOCKERBUDDY_BIN))

	resposta := readLine(Tr("\nInstalar o DockerBuddy? [S/n]: ", "\nInstall DockerBuddy? [Y/n]: "))
	ativado := Confirmou(resposta, true)
	if ativado {
		PrintOk(Tr("DockerBuddy será instalado", "DockerBuddy will be installed"))
	} else {
		PrintInfo(Tr("DockerBuddy não será instalado", "DockerBuddy will not be installed"))
	}
	Registrar("dockerbuddy", "enabled", ativado)
	return ativado
}

// instalarDockerBuddy baixa o DockerBuddy e cria o comando. Falhas aqui não
// interrompem a instalação: é opcional e os serviços já subiram.
func instalarDockerBuddy() {
	PrintSection(Tr("Instalando o DockerBuddy...", "Installing DockerBuddy..."))

	if DRY_RUN {
		if existe(DOCKERBUDDY_DIR) {
			PrintInfo(Tr("[dry-run] manteria a pasta existente ", "[dry-run] would keep the existing folder ") + DOCKERBUDDY_DIR)
		} else {
			PrintInfo(fmt.Sprintf(Tr("[dry-run] baixaria %s para %s", "[dry-run] would download %s to %s"), DOCKERBUDDY_URL, DOCKERBUDDY_DIR))
		}
		PrintInfo(fmt.Sprintf(Tr("[dry-run] criaria o link %s -> %s", "[dry-run] would create the link %s -> %s"), DOCKERBUDDY_BIN, dockerbuddyScript()))
		return
	}

	// Não sobrescreve uma pasta existente: os .txt podem ter sido editados à mão.
	if existe(DOCKERBUDDY_DIR) {
		PrintOk(DOCKERBUDDY_DIR + Tr(" já existe, mantido", " already exists, kept"))
	} else if err := baixarDockerBuddy(); err != nil {
		PrintWarning(Tr("Não foi possível baixar o DockerBuddy: ", "Could not download DockerBuddy: ") + err.Error())
		Registrar("dockerbuddy", "result", "failed", "error", err.Error())
		return
	} else {
		PrintOk(Tr("DockerBuddy baixado em ", "DockerBuddy downloaded to ") + DOCKERBUDDY_DIR)
	}

	if err := linkarDockerBuddy(); err != nil {
		PrintWarning(err.Error())
		Registrar("dockerbuddy", "result", "failed", "error", err.Error())
		return
	}

	PrintOk(Tr("DockerBuddy pronto. Digite 'dockerbuddy' no terminal para abrir a colinha.",
		"DockerBuddy ready. Type 'dockerbuddy' in the terminal to open the cheat sheet."))
	Registrar("dockerbuddy", "result", "installed")
}

// linkarDockerBuddy cria DOCKERBUDDY_BIN apontando para o script. Se já
// existir um arquivo lá que não seja esse link, ele é mantido.
func linkarDockerBuddy() error {
	alvo := dockerbuddyScript()
	if info, err := os.Lstat(DOCKERBUDDY_BIN); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if atual, _ := os.Readlink(DOCKERBUDDY_BIN); atual == alvo {
				PrintOk(DOCKERBUDDY_BIN + Tr(" já existe, mantido", " already exists, kept"))
				return nil
			}
		}
		return fmt.Errorf(Tr("%s já existe e não aponta para %s; mantido sem alterar",
			"%s already exists and does not point to %s; left unchanged"), DOCKERBUDDY_BIN, alvo)
	}
	if err := os.MkdirAll(filepath.Dir(DOCKERBUDDY_BIN), 0o755); err != nil {
		return fmt.Errorf(Tr("não foi possível criar %s: %w", "could not create %s: %w"), filepath.Dir(DOCKERBUDDY_BIN), err)
	}
	if err := os.Symlink(alvo, DOCKERBUDDY_BIN); err != nil {
		return fmt.Errorf(Tr("não foi possível criar o link %s: %w", "could not create the link %s: %w"), DOCKERBUDDY_BIN, err)
	}
	PrintOk(fmt.Sprintf(Tr("Link criado: %s -> %s", "Link created: %s -> %s"), DOCKERBUDDY_BIN, alvo))
	return nil
}

// baixarDockerBuddy baixa o .tar.gz do repositório e extrai só os arquivos
// conhecidos numa pasta temporária, que depois é renomeada para
// DOCKERBUDDY_DIR. Assim uma falha no meio não deixa uma instalação pela metade.
func baixarDockerBuddy() error {
	PrintInfo(Tr("Baixando ", "Downloading ") + DOCKERBUDDY_URL + "...")
	cliente := &http.Client{Timeout: 60 * time.Second}
	resp, err := cliente.Get(DOCKERBUDDY_URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(Tr("download respondeu %s", "download returned %s"), resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(DOCKERBUDDY_DIR), 0o755); err != nil {
		return err
	}
	tmp := DOCKERBUDDY_DIR + ".novo"
	os.RemoveAll(tmp)
	if err := extrairDockerBuddy(io.LimitReader(resp.Body, dockerbuddyMaxBytes), tmp); err != nil {
		os.RemoveAll(tmp)
		return err
	}
	if err := os.Rename(tmp, DOCKERBUDDY_DIR); err != nil {
		os.RemoveAll(tmp)
		return err
	}
	return nil
}

// extrairDockerBuddy lê o tar.gz e grava em destino apenas dockerbuddy.sh,
// README.md, LICENSE e data/*.txt. Qualquer outro caminho é ignorado, o que
// também impede que o arquivo escreva fora de destino.
func extrairDockerBuddy(r io.Reader, destino string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf(Tr("arquivo baixado não é um .tar.gz válido: %w", "downloaded file is not a valid .tar.gz: %w"), err)
	}
	defer gz.Close()

	if err := os.MkdirAll(filepath.Join(destino, "data"), 0o755); err != nil {
		return err
	}

	temScript, temDados := false, false
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf(Tr("erro lendo o .tar.gz: %w", "error reading the .tar.gz: %w"), err)
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}

		// O GitHub coloca tudo dentro de "DockerBuddy-main/"; descarta esse prefixo.
		_, relativo, ok := strings.Cut(h.Name, "/")
		if !ok {
			continue
		}
		modo := os.FileMode(0o644)
		switch {
		case relativo == "dockerbuddy.sh":
			modo = 0o755
			temScript = true
		case relativo == "README.md" || relativo == "LICENSE":
		case path.Dir(relativo) == "data" && strings.HasSuffix(relativo, ".txt"):
			temDados = true
		default:
			continue
		}

		f, err := os.OpenFile(filepath.Join(destino, filepath.FromSlash(relativo)),
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC, modo)
		if err != nil {
			return err
		}
		_, errCopia := io.Copy(f, tr)
		errFechar := f.Close()
		if errCopia != nil {
			return errCopia
		}
		if errFechar != nil {
			return errFechar
		}
	}

	if !temScript || !temDados {
		return errors.New(Tr("o arquivo baixado não tem dockerbuddy.sh e data/*.txt", "the downloaded file lacks dockerbuddy.sh and data/*.txt"))
	}
	return nil
}

// desinstalarDockerBuddy remove o link e a pasta, se o usuário confirmar.
func desinstalarDockerBuddy() {
	if !dockerbuddyInstalado() {
		return
	}
	PrintSection(Tr("DockerBuddy (colinha de comandos Docker)", "DockerBuddy (Docker command cheat sheet)"))
	resposta := readLine(Tr("Também desinstalar o DockerBuddy? [s/N]: ", "Also uninstall DockerBuddy? [y/N]: "))
	if !Confirmou(resposta, false) {
		PrintInfo(Tr("DockerBuddy mantido.", "DockerBuddy kept."))
		Registrar("dockerbuddy", "uninstall", "kept")
		return
	}

	if DRY_RUN {
		PrintInfo(fmt.Sprintf(Tr("[dry-run] removeria %s e %s", "[dry-run] would remove %s and %s"), DOCKERBUDDY_BIN, DOCKERBUDDY_DIR))
		return
	}

	// Só remove o link se ele for nosso; um arquivo com o mesmo nome criado
	// por outra pessoa fica.
	if info, err := os.Lstat(DOCKERBUDDY_BIN); err == nil {
		atual, _ := os.Readlink(DOCKERBUDDY_BIN)
		if info.Mode()&os.ModeSymlink != 0 && atual == dockerbuddyScript() {
			if err := os.Remove(DOCKERBUDDY_BIN); err != nil {
				PrintWarning(DOCKERBUDDY_BIN + Tr(" não pôde ser removido: ", " could not be removed: ") + err.Error())
			} else {
				PrintOk(DOCKERBUDDY_BIN + Tr(" removido", " removed"))
			}
		} else {
			PrintInfo(DOCKERBUDDY_BIN + Tr(" não aponta para o DockerBuddy instalado aqui; mantido",
				" does not point to the DockerBuddy installed here; kept"))
		}
	}
	if existe(DOCKERBUDDY_DIR) {
		if err := os.RemoveAll(DOCKERBUDDY_DIR); err != nil {
			PrintWarning(DOCKERBUDDY_DIR + Tr(" não pôde ser removido: ", " could not be removed: ") + err.Error())
		} else {
			PrintOk(DOCKERBUDDY_DIR + Tr(" removido", " removed"))
		}
	}
	Registrar("dockerbuddy", "uninstall", "removed")
}
