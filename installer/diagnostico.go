package installer

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Diagnostico struct {
	Servico  string
	Problema string
	Correcao string
	chave    string
	corrigir func() error
}

type regraDiagnostico struct {
	nome         string
	padrao       *regexp.Regexp
	diagnosticar func(servico, composePath string, achado []string) Diagnostico
}

var regrasDiagnostico = []regraDiagnostico{
	{"env_ausente", regexp.MustCompile(`env file (\S+) not found`), diagnosticarEnvAusente},
	{"variavel_ausente", regexp.MustCompile(`The \\?"(\w+)\\?" variable is not set`), diagnosticarVariavelAusente},
}

var exemplosEnv = []string{"env-example.env", ".env.example", "env.example", ".env.sample"}

var (
	rePlaceholder  = regexp.MustCompile(`^[A-ZÀ-Ú]+( [A-ZÀ-Ú]+)+$`)
	reChaveSecreta = regexp.MustCompile(`PASSWORD|SECRET|TOKEN|_KEY$`)
)

func Diagnosticar(servico, composePath, saida string) []Diagnostico {
	var achados []Diagnostico
	vistos := map[string]bool{}
	for _, r := range regrasDiagnostico {
		m := r.padrao.FindStringSubmatch(saida)
		if m == nil {
			continue
		}
		d := r.diagnosticar(servico, composePath, m)
		if vistos[d.chave] {
			continue
		}
		vistos[d.chave] = true
		slog.Info("Diagnóstico encontrado", "servico", servico, "regra", r.nome)
		achados = append(achados, d)
	}
	return achados
}

func tentarCorrigir(falhas []falhaServico, composeFiles map[string]string, repetir func(nome string) (string, error)) []falhaServico {
	var restantes []falhaServico
	for _, f := range falhas {
		corrigiu := false
		for _, d := range Diagnosticar(f.nome, composeFiles[f.nome], f.erro) {
			PrintWarning(fmt.Sprintf("%s: %s.", d.Servico, d.Problema))
			if d.corrigir == nil {
				continue
			}
			if strings.ToLower(readLine(fmt.Sprintf("Posso %s? [s/N]: ", d.Correcao))) != "s" {
				Registrar("correcao", "servico", d.Servico, "problema", d.Problema, "resultado", "recusada")
				continue
			}
			if err := d.corrigir(); err != nil {
				PrintFail(fmt.Sprintf("%s: não foi possível corrigir: %v", d.Servico, err))
				Registrar("correcao", "servico", d.Servico, "problema", d.Problema, "resultado", "falhou", "erro", err.Error())
				continue
			}
			Registrar("correcao", "servico", d.Servico, "problema", d.Problema, "resultado", "aplicada")
			corrigiu = true
		}

		if !corrigiu {
			restantes = append(restantes, f)
			continue
		}
		stderr, err := repetir(f.nome)
		if err != nil {
			restantes = append(restantes, falhaServico{f.nome, stderr})
			continue
		}
		PrintOk(f.nome + " corrigido")
	}
	return restantes
}

func encontrarExemploEnv(dir string) string {
	for _, nome := range exemplosEnv {
		if caminho := filepath.Join(dir, nome); existe(caminho) {
			return caminho
		}
	}
	return ""
}

func diagnosticarEnvAusente(servico, composePath string, achado []string) Diagnostico {
	alvo := filepath.Clean(achado[1])
	d := Diagnostico{
		Servico:  servico,
		chave:    "env:" + alvo,
		Problema: fmt.Sprintf("o arquivo %s não existe, mas o docker-compose depende dele (env_file)", alvo),
	}

	dir := filepath.Dir(composePath)
	if resolver(filepath.Dir(alvo)) != resolver(dir) {
		return d
	}
	exemplo := encontrarExemploEnv(dir)
	if exemplo == "" {
		return d
	}

	d.Correcao = fmt.Sprintf("criar %s a partir de %s, gerando as senhas e perguntando o resto",
		filepath.Base(alvo), filepath.Base(exemplo))
	d.corrigir = func() error { return criarEnvDeExemplo(exemplo, alvo, servico) }
	return d
}

func diagnosticarVariavelAusente(servico, composePath string, achado []string) Diagnostico {
	alvo := filepath.Join(filepath.Dir(composePath), ".env")
	if !existe(alvo) {
		return diagnosticarEnvAusente(servico, composePath, []string{"", alvo})
	}
	return Diagnostico{
		Servico:  servico,
		chave:    "variavel:" + achado[1],
		Problema: fmt.Sprintf("a variável %s não está definida em %s", achado[1], alvo),
	}
}

func valorParaChave(chave, servico string) (valor string, gerado bool) {
	switch {
	case chave == "APP_KEY":
		bytes := make([]byte, 32)
		rand.Read(bytes)
		return "base64:" + base64.StdEncoding.EncodeToString(bytes), true
	case reChaveSecreta.MatchString(chave):
		return rand.Text(), true
	}

	padrao, _, _ := strings.Cut(servico, "-")
	if resposta := readLine(fmt.Sprintf("Valor para %s (%s) [%s]: ", chave, servico, padrao)); resposta != "" {
		return resposta, false
	}
	return padrao, false
}

func criarEnvDeExemplo(exemplo, alvo, servico string) error {
	dados, err := os.ReadFile(exemplo)
	if err != nil {
		return err
	}

	linhas := strings.Split(string(dados), "\n")
	var gerados []string
	for i, linha := range linhas {
		chave, valor, ok := strings.Cut(linha, "=")
		chave = strings.TrimSpace(chave)
		if !ok || chave == "" || strings.HasPrefix(chave, "#") {
			continue
		}
		valor = strings.TrimSpace(valor)
		if valor != "" && !rePlaceholder.MatchString(valor) {
			continue
		}
		novo, gerado := valorParaChave(chave, servico)
		if gerado {
			gerados = append(gerados, chave)
		}
		linhas[i] = chave + "=" + novo
	}

	if DRY_RUN {
		PrintInfo(fmt.Sprintf("[dry-run] criaria %s a partir de %s", alvo, filepath.Base(exemplo)))
		return nil
	}

	if err := os.WriteFile(alvo, []byte(strings.Join(linhas, "\n")), 0o600); err != nil {
		return err
	}
	if owner := detectRealUser(); owner != nil {
		os.Chown(alvo, owner.uid, owner.gid)
	}
	slog.Info("Arquivo .env criado a partir do exemplo", "servico", servico, "arquivo", alvo, "gerados", gerados)

	PrintOk(alvo + " criado")
	if len(gerados) > 0 {
		PrintInfo(fmt.Sprintf("Gerados automaticamente: %s. Consulte os valores em %s.", strings.Join(gerados, ", "), alvo))
	}
	return nil
}
