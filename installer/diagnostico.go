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
	{"missing_env_file", regexp.MustCompile(`env file (\S+) not found`), diagnosticarEnvAusente},
	{"missing_variable", regexp.MustCompile(`The \\?"(\w+)\\?" variable is not set`), diagnosticarVariavelAusente},
	{"undefined_service", regexp.MustCompile(`service \\?"([\w.-]+)\\?" depends on undefined service \\?"([\w.-]+)\\?"`), diagnosticarServicoIndefinido},
}

const maxRodadasCorrecao = 3

var exemplosEnv = []string{"env-example.env", ".env.example", "env.example", ".env.sample"}

var (
	rePlaceholder  = regexp.MustCompile(`^[A-ZÀ-Ú]+( [A-ZÀ-Ú]+)+$`)
	reChaveSecreta = regexp.MustCompile(`PASSWORD|SECRET|TOKEN|_KEY$`)
)

func Diagnosticar(servico, composePath, saida string) []Diagnostico {
	var achados []Diagnostico
	vistos := map[string]bool{}
	for _, r := range regrasDiagnostico {
		for _, m := range r.padrao.FindAllStringSubmatch(saida, -1) {
			d := r.diagnosticar(servico, composePath, m)
			if vistos[d.chave] {
				continue
			}
			vistos[d.chave] = true
			slog.Info("Diagnosis found", "service", servico, "rule", r.nome)
			achados = append(achados, d)
		}
	}
	return achados
}

// Repete diagnóstico + correção enquanto houver progresso: corrigir uma coisa
// (ex.: criar o .env) costuma revelar o próximo erro (ex.: variável faltando).
func tentarCorrigir(falhas []falhaServico, composeFiles map[string]string, repetir func(nome string) (string, error)) []falhaServico {
	var restantes []falhaServico
	for _, f := range falhas {
		corrigiuAlgo, resolvido := false, false
		for rodada := 0; rodada < maxRodadasCorrecao; rodada++ {
			if !aplicarCorrecoes(f.nome, composeFiles[f.nome], f.erro) {
				break
			}
			corrigiuAlgo = true
			stderr, err := repetir(f.nome)
			if err == nil {
				resolvido = true
				break
			}
			f.erro = stderr
		}

		if resolvido {
			PrintOk(f.nome + Tr(" corrigido", " fixed"))
			continue
		}
		if corrigiuAlgo {
			PrintWarning(fmt.Sprintf(Tr("%s: as correções foram aplicadas, mas ainda há um erro que não é do .env (veja abaixo).",
				"%s: the fixes were applied, but there is still an error unrelated to the .env (see below)."), f.nome))
		}
		restantes = append(restantes, f)
	}
	return restantes
}

func aplicarCorrecoes(servico, composePath, saida string) bool {
	corrigiu := false
	for _, d := range Diagnosticar(servico, composePath, saida) {
		PrintWarning(fmt.Sprintf("%s: %s.", d.Servico, d.Problema))
		if d.corrigir == nil {
			continue
		}
		if !Confirmou(readLine(fmt.Sprintf(Tr("Posso %s? [s/N]: ", "May I %s? [y/N]: "), d.Correcao)), false) {
			Registrar("fix", "service", d.Servico, "problem", d.chave, "result", "declined")
			continue
		}
		if err := d.corrigir(); err != nil {
			PrintFail(fmt.Sprintf(Tr("%s: não foi possível corrigir: %v", "%s: could not fix: %v"), d.Servico, err))
			Registrar("fix", "service", d.Servico, "problem", d.chave, "result", "failed", "error", err.Error())
			continue
		}
		Registrar("fix", "service", d.Servico, "problem", d.chave, "result", "applied")
		// em dry-run nada foi escrito, então repetir o comando só daria o mesmo erro
		if !DRY_RUN {
			corrigiu = true
		}
	}
	return corrigiu
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
		Servico: servico,
		chave:   "env:" + alvo,
		Problema: fmt.Sprintf(Tr("o arquivo %s não existe, mas o docker-compose depende dele (env_file)",
			"the file %s does not exist, but the docker-compose depends on it (env_file)"), alvo),
	}

	dir := filepath.Dir(composePath)
	if resolver(filepath.Dir(alvo)) != resolver(dir) {
		return d
	}
	exemplo := encontrarExemploEnv(dir)
	if exemplo == "" {
		return d
	}

	d.Correcao = fmt.Sprintf(Tr("criar %s a partir de %s, gerando as senhas e perguntando o resto",
		"create %s from %s, generating the passwords and asking for the rest"),
		filepath.Base(alvo), filepath.Base(exemplo))
	d.corrigir = func() error { return criarEnvDeExemplo(exemplo, alvo, servico) }
	return d
}

func diagnosticarVariavelAusente(servico, composePath string, achado []string) Diagnostico {
	alvo := filepath.Join(filepath.Dir(composePath), ".env")
	if !existe(alvo) {
		return diagnosticarEnvAusente(servico, composePath, []string{"", alvo})
	}
	chave := achado[1]
	exemplo := encontrarExemploEnv(filepath.Dir(composePath))
	return Diagnostico{
		Servico:  servico,
		chave:    "variable:" + chave,
		Problema: fmt.Sprintf(Tr("a variável %s não está definida em %s", "the variable %s is not set in %s"), chave, alvo),
		Correcao: fmt.Sprintf(Tr("adicionar %s ao %s", "add %s to %s"), chave, filepath.Base(alvo)),
		corrigir: func() error { return adicionarVariavel(alvo, chave, servico, exemplo) },
	}
}

func diagnosticarServicoIndefinido(servico, composePath string, achado []string) Diagnostico {
	return Diagnostico{
		Servico: servico,
		chave:   "service:" + achado[2],
		Problema: fmt.Sprintf(Tr("o serviço %s depende de %s, que não existe em %s — o arquivo parece incompleto, compare com o compose oficial",
			"the service %s depends on %s, which does not exist in %s — the file looks incomplete, compare it with the official compose"),
			achado[1], achado[2], composePath),
	}
}

// Valor já preenchido no exemplo (não placeholder) para a chave, se houver.
func valorNoExemplo(exemplo, chave string) (string, bool) {
	if exemplo == "" {
		return "", false
	}
	dados, err := os.ReadFile(exemplo)
	if err != nil {
		return "", false
	}
	for _, linha := range strings.Split(string(dados), "\n") {
		k, v, ok := strings.Cut(linha, "=")
		if !ok || strings.TrimSpace(k) != chave {
			continue
		}
		v = strings.TrimSpace(v)
		if v != "" && !rePlaceholder.MatchString(v) {
			return v, true
		}
	}
	return "", false
}

func adicionarVariavel(alvo, chave, servico, exemplo string) error {
	valor, ok := valorNoExemplo(exemplo, chave)
	gerado := false
	if !ok {
		valor, gerado = valorParaChave(chave, servico)
	}

	if DRY_RUN {
		PrintInfo(fmt.Sprintf(Tr("[dry-run] adicionaria %s ao %s", "[dry-run] would add %s to %s"), chave, alvo))
		return nil
	}

	atual, err := os.ReadFile(alvo)
	if err != nil {
		return err
	}
	linha := chave + "=" + valor + "\n"
	if len(atual) > 0 && !strings.HasSuffix(string(atual), "\n") {
		linha = "\n" + linha
	}
	arq, err := os.OpenFile(alvo, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer arq.Close()
	if _, err := arq.WriteString(linha); err != nil {
		return err
	}
	slog.Info("Variable added to .env", "service", servico, "file", alvo, "key", chave, "generated", gerado)

	PrintOk(fmt.Sprintf(Tr("%s adicionada a %s", "%s added to %s"), chave, alvo))
	if gerado {
		PrintInfo(fmt.Sprintf(Tr("%s gerada automaticamente. Consulte o valor em %s.", "%s generated automatically. See the value in %s."), chave, alvo))
	}
	return nil
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
	if resposta := readLine(fmt.Sprintf(Tr("Valor para %s (%s) [%s]: ", "Value for %s (%s) [%s]: "), chave, servico, padrao)); resposta != "" {
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
		PrintInfo(fmt.Sprintf(Tr("[dry-run] criaria %s a partir de %s", "[dry-run] would create %s from %s"), alvo, filepath.Base(exemplo)))
		return nil
	}

	// O_EXCL: nunca sobrescreve um .env que o usuário já tenha criado nesse meio-tempo
	arq, err := os.OpenFile(alvo, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, err = arq.WriteString(strings.Join(linhas, "\n"))
	if fechar := arq.Close(); err == nil {
		err = fechar
	}
	if err != nil {
		return err
	}
	if owner := detectRealUser(); owner != nil {
		os.Chown(alvo, owner.uid, owner.gid)
	}
	slog.Info(".env file created from the example", "service", servico, "file", alvo, "generated", gerados)

	PrintOk(alvo + Tr(" criado", " created"))
	if len(gerados) > 0 {
		PrintInfo(fmt.Sprintf(Tr("Gerados automaticamente: %s. Consulte os valores em %s.", "Generated automatically: %s. See the values in %s."), strings.Join(gerados, ", "), alvo))
	}
	return nil
}
