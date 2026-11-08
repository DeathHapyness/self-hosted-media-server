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
	PrintSection("Colinha de comandos Docker (DockerBuddy)")
	PrintInfo("O DockerBuddy é uma colinha de comandos Docker no terminal. Você digita " +
		"'dockerbuddy', escolhe um número e ele mostra os comandos mais usados, cada um " +
		"com uma explicação curta: containers, imagens, volumes, redes, Docker Compose " +
		"e filtros do 'docker ps'.")
	PrintInfo("Serve para o dia a dia do servidor: ver logs, reiniciar um serviço, " +
		"limpar imagens velhas, sem precisar procurar o comando na internet.")
	PrintInfo(fmt.Sprintf("Não é container nem serviço: não abre porta e não fica rodando. "+
		"São só um script e alguns arquivos de texto em %s, com o comando %s. "+
		"Detalhes em docs/dockerbuddy.md.", DOCKERBUDDY_DIR, DOCKERBUDDY_BIN))

	resposta := strings.ToLower(readLine("\nInstalar o DockerBuddy? [S/n]: "))
	ativado := resposta != "n" && resposta != "nao" && resposta != "não"
	if ativado {
		PrintOk("DockerBuddy será instalado")
	} else {
		PrintInfo("DockerBuddy não será instalado")
	}
	Registrar("dockerbuddy", "ativado", ativado)
	return ativado
}

// instalarDockerBuddy baixa o DockerBuddy e cria o comando. Falhas aqui não
// interrompem a instalação: é opcional e os serviços já subiram.
func instalarDockerBuddy() {
	PrintSection("Instalando o DockerBuddy...")

	if DRY_RUN {
		if existe(DOCKERBUDDY_DIR) {
			PrintInfo("[dry-run] manteria a pasta existente " + DOCKERBUDDY_DIR)
		} else {
			PrintInfo(fmt.Sprintf("[dry-run] baixaria %s para %s", DOCKERBUDDY_URL, DOCKERBUDDY_DIR))
		}
		PrintInfo(fmt.Sprintf("[dry-run] criaria o link %s -> %s", DOCKERBUDDY_BIN, dockerbuddyScript()))
		return
	}

	// Não sobrescreve uma pasta existente: os .txt podem ter sido editados à mão.
	if existe(DOCKERBUDDY_DIR) {
		PrintOk(DOCKERBUDDY_DIR + " já existe, mantido")
	} else if err := baixarDockerBuddy(); err != nil {
		PrintWarning("Não foi possível baixar o DockerBuddy: " + err.Error())
		Registrar("dockerbuddy", "resultado", "falhou", "erro", err.Error())
		return
	} else {
		PrintOk("DockerBuddy baixado em " + DOCKERBUDDY_DIR)
	}

	if err := linkarDockerBuddy(); err != nil {
		PrintWarning(err.Error())
		Registrar("dockerbuddy", "resultado", "falhou", "erro", err.Error())
		return
	}

	PrintOk("DockerBuddy pronto. Digite 'dockerbuddy' no terminal para abrir a colinha.")
	Registrar("dockerbuddy", "resultado", "instalado")
}

// linkarDockerBuddy cria DOCKERBUDDY_BIN apontando para o script. Se já
// existir um arquivo lá que não seja esse link, ele é mantido.
func linkarDockerBuddy() error {
	alvo := dockerbuddyScript()
	if info, err := os.Lstat(DOCKERBUDDY_BIN); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if atual, _ := os.Readlink(DOCKERBUDDY_BIN); atual == alvo {
				PrintOk(DOCKERBUDDY_BIN + " já existe, mantido")
				return nil
			}
		}
		return fmt.Errorf("%s já existe e não aponta para %s; mantido sem alterar", DOCKERBUDDY_BIN, alvo)
	}
	if err := os.MkdirAll(filepath.Dir(DOCKERBUDDY_BIN), 0o755); err != nil {
		return fmt.Errorf("não foi possível criar %s: %w", filepath.Dir(DOCKERBUDDY_BIN), err)
	}
	if err := os.Symlink(alvo, DOCKERBUDDY_BIN); err != nil {
		return fmt.Errorf("não foi possível criar o link %s: %w", DOCKERBUDDY_BIN, err)
	}
	PrintOk(fmt.Sprintf("Link criado: %s -> %s", DOCKERBUDDY_BIN, alvo))
	return nil
}

// baixarDockerBuddy baixa o .tar.gz do repositório e extrai só os arquivos
// conhecidos numa pasta temporária, que depois é renomeada para
// DOCKERBUDDY_DIR. Assim uma falha no meio não deixa uma instalação pela metade.
func baixarDockerBuddy() error {
	PrintInfo("Baixando " + DOCKERBUDDY_URL + "...")
	cliente := &http.Client{Timeout: 60 * time.Second}
	resp, err := cliente.Get(DOCKERBUDDY_URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download respondeu %s", resp.Status)
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
		return fmt.Errorf("arquivo baixado não é um .tar.gz válido: %w", err)
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
			return fmt.Errorf("erro lendo o .tar.gz: %w", err)
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
		return errors.New("o arquivo baixado não tem dockerbuddy.sh e data/*.txt")
	}
	return nil
}

// desinstalarDockerBuddy remove o link e a pasta, se o usuário confirmar.
func desinstalarDockerBuddy() {
	if !dockerbuddyInstalado() {
		return
	}
	PrintSection("DockerBuddy (colinha de comandos Docker)")
	resposta := strings.ToLower(readLine("Também desinstalar o DockerBuddy? [s/N]: "))
	if resposta != "s" {
		PrintInfo("DockerBuddy mantido.")
		Registrar("dockerbuddy", "desinstalacao", "mantido")
		return
	}

	if DRY_RUN {
		PrintInfo(fmt.Sprintf("[dry-run] removeria %s e %s", DOCKERBUDDY_BIN, DOCKERBUDDY_DIR))
		return
	}

	// Só remove o link se ele for nosso; um arquivo com o mesmo nome criado
	// por outra pessoa fica.
	if info, err := os.Lstat(DOCKERBUDDY_BIN); err == nil {
		atual, _ := os.Readlink(DOCKERBUDDY_BIN)
		if info.Mode()&os.ModeSymlink != 0 && atual == dockerbuddyScript() {
			if err := os.Remove(DOCKERBUDDY_BIN); err != nil {
				PrintWarning(DOCKERBUDDY_BIN + " não pôde ser removido: " + err.Error())
			} else {
				PrintOk(DOCKERBUDDY_BIN + " removido")
			}
		} else {
			PrintInfo(DOCKERBUDDY_BIN + " não aponta para o DockerBuddy instalado aqui; mantido")
		}
	}
	if existe(DOCKERBUDDY_DIR) {
		if err := os.RemoveAll(DOCKERBUDDY_DIR); err != nil {
			PrintWarning(DOCKERBUDDY_DIR + " não pôde ser removido: " + err.Error())
		} else {
			PrintOk(DOCKERBUDDY_DIR + " removido")
		}
	}
	Registrar("dockerbuddy", "desinstalacao", "removido")
}
