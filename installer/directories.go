package installer

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type dono struct {
	uid, gid int
}

func (d *dono) LogValue() slog.Value {
	if d == nil {
		return slog.StringValue("nenhum")
	}
	return slog.StringValue(fmt.Sprintf("%d:%d", d.uid, d.gid))
}

func envPath() string {
	return filepath.Join(repoRoot(), ".env")
}

func lerEnv(caminho string) map[string]string {
	valores := map[string]string{}
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return valores
	}
	for _, linha := range strings.Split(string(dados), "\n") {
		linha = strings.TrimSpace(linha)
		if linha == "" || strings.HasPrefix(linha, "#") || !strings.Contains(linha, "=") {
			continue
		}
		chave, valor, _ := strings.Cut(linha, "=")
		valores[strings.TrimSpace(chave)] = strings.TrimSpace(valor)
	}
	return valores
}

func detectRealUser() *dono {
	sudoUser := os.Getenv("SUDO_USER")
	if sudoUser == "" {
		return nil
	}
	u, err := user.Lookup(sudoUser)
	if err != nil {
		return nil
	}
	uid, err1 := strconv.Atoi(u.Uid)
	gid, err2 := strconv.Atoi(u.Gid)
	if err1 != nil || err2 != nil {
		return nil
	}
	return &dono{uid, gid}
}

func expandirHome(caminho string) string {
	if caminho != "~" && !strings.HasPrefix(caminho, "~/") {
		return caminho
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return caminho
	}
	return filepath.Join(home, caminho[1:])
}

func perguntarCaminho(rotulo, padrao string) string {
	for {
		resposta := readLine(fmt.Sprintf("%s [%s]: ", rotulo, padrao))
		if resposta == "" {
			return padrao
		}
		caminho := expandirHome(resposta)
		if filepath.IsAbs(caminho) {
			return caminho
		}
		PrintWarning("Use um caminho absoluto, começando com /.")
	}
}

func LoadEnvIntoState() {
	valores := lerEnv(envPath())
	if valores["DATA_DIR"] != "" {
		BASE_DIR = valores["DATA_DIR"]
	}
	if valores["MEDIA_DIR"] != "" {
		MEDIA_DIR = valores["MEDIA_DIR"]
	}
}

func resolvePuidPgid() *dono {
	valores := lerEnv(envPath())
	uid, err1 := strconv.Atoi(valores["PUID"])
	gid, err2 := strconv.Atoi(valores["PGID"])
	if err1 != nil || err2 != nil {
		return nil
	}
	return &dono{uid, gid}
}

func setupEnv() error {
	env := envPath()
	example := filepath.Join(repoRoot(), ".env.example")

	if _, err := os.Stat(env); err == nil {
		LoadEnvIntoState()
		slog.Info("Caminhos lidos do .env", "dados", BASE_DIR, "midia", MEDIA_DIR)
		Registrar("caminhos", "origem", ".env", "dados", BASE_DIR, "midia", MEDIA_DIR)
		PrintOk(".env already present")
		PrintInfo("dados/configs: " + BASE_DIR)
		PrintInfo("biblioteca de mídia: " + MEDIA_DIR)
		return nil
	}

	PrintSection("Onde este servidor vai guardar os arquivos?")
	PrintInfo("Os padrões servem para a maioria dos casos — Enter aceita cada um.")
	base := perguntarCaminho("Dados e configs dos serviços", DefaultBaseDir)
	media := perguntarCaminho("Biblioteca de mídia", DefaultMediaDir)
	BASE_DIR = base
	MEDIA_DIR = media
	slog.Info("Caminhos escolhidos pelo usuario", "dados", base, "midia", media)
	Registrar("caminhos", "origem", "usuario", "dados", base, "midia", media)

	owner := detectRealUser()
	slog.Info("PUID/PGID detectados", "owner", owner)

	if DRY_RUN {
		PrintInfo(fmt.Sprintf("[dry-run] criaria .env com DATA_DIR=%s e MEDIA_DIR=%s", base, media))
		if owner != nil {
			PrintInfo(fmt.Sprintf("[dry-run] gravaria PUID=%d PGID=%d", owner.uid, owner.gid))
		}
		return nil
	}

	conteudo := "PUID=1000\nPGID=1000\nTZ=America/Sao_Paulo\n"
	if dados, err := os.ReadFile(example); err == nil {
		conteudo = string(dados)
	}

	chaves := []string{"DATA_DIR", "MEDIA_DIR"}
	valores := map[string]string{"DATA_DIR": base, "MEDIA_DIR": media}
	if owner != nil {
		chaves = append(chaves, "PUID", "PGID")
		valores["PUID"] = strconv.Itoa(owner.uid)
		valores["PGID"] = strconv.Itoa(owner.gid)
	}

	for _, chave := range chaves {
		linha := chave + "=" + valores[chave]
		re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(chave) + `=.*$`)
		if re.MatchString(conteudo) {
			conteudo = re.ReplaceAllLiteralString(conteudo, linha)
		} else {
			if !strings.HasSuffix(conteudo, "\n") {
				conteudo += "\n"
			}
			conteudo += linha + "\n"
		}
	}

	if err := os.WriteFile(env, []byte(conteudo), 0o644); err != nil {
		return fmt.Errorf("Não foi possível criar %s: %v", env, err)
	}
	slog.Info("Arquivo .env criado", "caminho", env)
	PrintOk(fmt.Sprintf(".env created (DATA_DIR=%s, MEDIA_DIR=%s)", base, media))
	if owner != nil {
		PrintOk(fmt.Sprintf("PUID=%d PGID=%d detectados a partir de SUDO_USER", owner.uid, owner.gid))
	} else {
		PrintWarning("SUDO_USER não definido — PUID/PGID ficaram com os valores do exemplo. " +
			"Confira com 'id -u' e 'id -g' antes de subir os serviços.")
	}
	PrintWarning("Revise o .env criado e preencha os valores necessários antes de subir os serviços.")
	return nil
}

func existe(caminho string) bool {
	_, err := os.Stat(caminho)
	return err == nil
}

func createDirectories() error {
	PrintSection("Creating directories...")
	dirs := ServiceDirs()
	nomes := make([]string, 0, len(dirs))
	for nome := range dirs {
		nomes = append(nomes, nome)
	}
	slices.SortFunc(nomes, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	for _, nome := range nomes {
		criouAlgum := false
		for _, d := range dirs[nome] {
			existia := existe(d)
			if DRY_RUN {
				if !existia {
					criouAlgum = true
					PrintInfo("[dry-run] criaria " + d)
				}
				continue
			}
			if err := os.MkdirAll(d, 0o755); err != nil {
				return fmt.Errorf("Não foi possível criar %s: %v", d, err)
			}
			if !existia {
				criouAlgum = true
			}
		}
		if criouAlgum {
			PrintOk(nome + " directories")
		} else {
			PrintOk(nome + " directories (already existed)")
		}
	}

	for _, sub := range MEDIA_SUBDIRS {
		caminho := filepath.Join(MEDIA_DIR, sub)
		existia := existe(caminho)
		if DRY_RUN {
			if !existia {
				PrintInfo("[dry-run] criaria " + caminho)
			}
			continue
		}
		if err := os.MkdirAll(caminho, 0o755); err != nil {
			return fmt.Errorf("Não foi possível criar %s: %v", caminho, err)
		}
		if !existia {
			PrintOk(caminho + " created")
		}
	}
	return nil
}

func configurePermissions() {
	PrintSection("Configuring permissions...")
	owner := resolvePuidPgid()
	alvos := writable_dirs()
	slog.Info("Ajustando permissoes", "total", len(alvos), "base", BASE_DIR, "owner", owner)

	if DRY_RUN {
		PrintInfo(fmt.Sprintf("[dry-run] ajustaria permissões de %d diretórios e de %s", len(alvos), MEDIA_DIR))
		PrintOk("Permissions configured (simulado)")
		return
	}

	for _, d := range alvos {
		if !existe(d) {
			continue
		}
		if err := os.Chmod(d, WRITABLE_MODE); err != nil {
			PrintWarning(fmt.Sprintf("Não foi possível ajustar permissões de %s: %v", d, err))
		}
		if owner != nil {
			if err := os.Chown(d, owner.uid, owner.gid); err != nil {
				PrintWarning(fmt.Sprintf("Não foi possível ajustar dono de %s para PUID/PGID do .env", d))
			}
		}
	}

	for _, d := range []string{MEDIA_DIR, filepath.Join(MEDIA_DIR, "downloads")} {
		if existe(d) {
			if err := os.Chmod(d, 0o775); err != nil {
				PrintWarning(fmt.Sprintf("Não foi possível ajustar permissões de %s: %v", d, err))
			}
		}
	}

	if owner != nil {
		PrintOk(fmt.Sprintf("Permissions configured (PUID=%d PGID=%d)", owner.uid, owner.gid))
	} else {
		PrintOk("Permissions configured")
		PrintWarning("Nenhum PUID/PGID encontrado em .env — mantendo dono atual dos " +
			"diretórios. Ajuste manualmente se os containers rodarem com outro " +
			"usuário.")
	}
	PrintInfo("Jellyfin e Navidrome devem montar a mídia como somente leitura (:ro) no docker-compose.yml")
}
