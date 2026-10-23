package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type uiFalsa struct{ respostas []string }

func (*uiFalsa) Cabecalho(string) {}
func (*uiFalsa) Secao(string)     {}
func (*uiFalsa) Ok(string)        {}
func (*uiFalsa) Falha(string)     {}
func (*uiFalsa) Aviso(string)     {}
func (*uiFalsa) Info(string)      {}
func (*uiFalsa) Perigo(string)    {}
func (*uiFalsa) Texto(string)     {}

func (u *uiFalsa) Perguntar(string) (string, error) {
	if len(u.respostas) == 0 {
		return "", nil
	}
	resposta := u.respostas[0]
	u.respostas = u.respostas[1:]
	return resposta, nil
}

func usarUI(t *testing.T, respostas ...string) {
	anterior := UI
	UI = &uiFalsa{respostas}
	t.Cleanup(func() { UI = anterior })
}

const exemploEnv = `# comentário que deve continuar igual
UPLOAD_LOCATION=./library
DB_PASSWORD=SENHA QUE VOCÊ DESEJA USAR PARA O BANCO DE DADOS
DB_USERNAME=USUARIO DO BANCO DE DADOS
APP_KEY=
IMMICH_VERSION=v3
`

func servicoDeTeste(t *testing.T, comExemplo bool) (dir, compose string) {
	dir = filepath.Join(t.TempDir(), "immich-app")
	os.Mkdir(dir, 0o755)
	compose = filepath.Join(dir, "docker-compose.yml")
	os.WriteFile(compose, []byte("services: {}\n"), 0o644)
	if comExemplo {
		os.WriteFile(filepath.Join(dir, "env-example.env"), []byte(exemploEnv), 0o644)
	}
	return dir, compose
}

func saidaEnvAusente(dir string) string {
	env := filepath.Join(dir, ".env")
	return `time="2026-10-23T00:33:08-03:00" level=warning msg="The \"KOPIA_SERVER_USERNAME\" variable is not set. Defaulting to a blank string."
env file ` + env + ` not found: stat ` + env + `: no such file or directory`
}

func TestDiagnosticarEnvAusenteSemDuplicar(t *testing.T) {
	dir, compose := servicoDeTeste(t, true)

	achados := Diagnosticar("immich-app", compose, saidaEnvAusente(dir))
	if len(achados) != 1 {
		t.Fatalf("esperava 1 diagnóstico, veio %d: %+v", len(achados), achados)
	}
	if achados[0].corrigir == nil {
		t.Fatal("esperava uma correção disponível")
	}
}

func TestDiagnosticarVariavelSemEnv(t *testing.T) {
	dir, compose := servicoDeTeste(t, true)
	saida := `time="2026-10-23T00:33:01-03:00" level=warning msg="The \"UPLOAD_LOCATION\" variable is not set. Defaulting to a blank string."
invalid spec: :/data: empty section between colons`

	achados := Diagnosticar("immich-app", compose, saida)
	if len(achados) != 1 || achados[0].corrigir == nil {
		t.Fatalf("esperava 1 diagnóstico com correção, veio %+v", achados)
	}
	if !strings.Contains(achados[0].Problema, filepath.Join(dir, ".env")) {
		t.Errorf("o problema deveria citar o .env: %q", achados[0].Problema)
	}
}

func TestDiagnosticarVariavelComEnvExistente(t *testing.T) {
	dir, compose := servicoDeTeste(t, true)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("OUTRA=1\n"), 0o600)

	achados := Diagnosticar("immich-app", compose, `The "UPLOAD_LOCATION" variable is not set. Defaulting to a blank string.`)
	if len(achados) != 1 || achados[0].corrigir != nil {
		t.Fatalf("esperava só o diagnóstico, sem correção: %+v", achados)
	}
}

func TestDiagnosticarSemExemploNaoCorrige(t *testing.T) {
	dir, compose := servicoDeTeste(t, false)

	achados := Diagnosticar("immich-app", compose, saidaEnvAusente(dir))
	if len(achados) != 1 || achados[0].corrigir != nil {
		t.Fatalf("sem arquivo de exemplo não deveria haver correção: %+v", achados)
	}
}

func TestDiagnosticarErroDesconhecido(t *testing.T) {
	_, compose := servicoDeTeste(t, true)

	if achados := Diagnosticar("immich-app", compose, "Error response from daemon: algo inesperado"); len(achados) != 0 {
		t.Fatalf("não deveria diagnosticar nada: %+v", achados)
	}
}

func TestCriarEnvDeExemplo(t *testing.T) {
	usarUI(t, "")
	dir, _ := servicoDeTeste(t, true)
	alvo := filepath.Join(dir, ".env")

	if err := criarEnvDeExemplo(filepath.Join(dir, "env-example.env"), alvo, "immich-app"); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(alvo)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("permissão %o, esperava 600", info.Mode().Perm())
	}

	dados, _ := os.ReadFile(alvo)
	if !strings.HasPrefix(string(dados), "# comentário que deve continuar igual\n") {
		t.Errorf("o comentário deveria ser preservado:\n%s", dados)
	}
	valores := lerEnv(alvo)
	if valores["UPLOAD_LOCATION"] != "./library" || valores["IMMICH_VERSION"] != "v3" {
		t.Errorf("valores já preenchidos não deveriam mudar: %v", valores)
	}
	if len(valores["DB_PASSWORD"]) < 20 || strings.Contains(valores["DB_PASSWORD"], " ") {
		t.Errorf("DB_PASSWORD deveria ser uma senha gerada: %q", valores["DB_PASSWORD"])
	}
	if !strings.HasPrefix(valores["APP_KEY"], "base64:") {
		t.Errorf("APP_KEY deveria ter o prefixo base64: %q", valores["APP_KEY"])
	}
	if valores["DB_USERNAME"] != "immich" {
		t.Errorf("DB_USERNAME deveria usar o padrão do serviço, veio %q", valores["DB_USERNAME"])
	}
}

func TestCriarEnvDeExemploUsaResposta(t *testing.T) {
	usarUI(t, "meu_usuario")
	dir, _ := servicoDeTeste(t, true)
	alvo := filepath.Join(dir, ".env")

	if err := criarEnvDeExemplo(filepath.Join(dir, "env-example.env"), alvo, "immich-app"); err != nil {
		t.Fatal(err)
	}
	if v := lerEnv(alvo)["DB_USERNAME"]; v != "meu_usuario" {
		t.Errorf("DB_USERNAME deveria ser a resposta do usuário, veio %q", v)
	}
}

func TestCriarEnvDeExemploDryRun(t *testing.T) {
	usarUI(t)
	DRY_RUN = true
	t.Cleanup(func() { DRY_RUN = false })
	dir, _ := servicoDeTeste(t, true)
	alvo := filepath.Join(dir, ".env")

	if err := criarEnvDeExemplo(filepath.Join(dir, "env-example.env"), alvo, "immich-app"); err != nil {
		t.Fatal(err)
	}
	if existe(alvo) {
		t.Error("no dry-run o .env não deveria ser criado")
	}
}

func TestTentarCorrigirRepeteSoQuemFoiCorrigido(t *testing.T) {
	usarUI(t, "s")
	dir, compose := servicoDeTeste(t, true)
	falhas := []falhaServico{
		{"immich-app", saidaEnvAusente(dir)},
		{"outro", "Error response from daemon: algo inesperado"},
	}

	var repetidos []string
	restantes := tentarCorrigir(falhas, map[string]string{"immich-app": compose, "outro": compose},
		func(nome string) (string, error) {
			repetidos = append(repetidos, nome)
			return "", nil
		})

	if len(repetidos) != 1 || repetidos[0] != "immich-app" {
		t.Errorf("só o immich-app deveria ser repetido, foram: %v", repetidos)
	}
	if len(restantes) != 1 || restantes[0].nome != "outro" {
		t.Errorf("só o 'outro' deveria continuar com falha: %+v", restantes)
	}
	if !existe(filepath.Join(dir, ".env")) {
		t.Error("o .env deveria ter sido criado")
	}
}

func TestTentarCorrigirRespeitaRecusa(t *testing.T) {
	usarUI(t, "n")
	dir, compose := servicoDeTeste(t, true)

	restantes := tentarCorrigir([]falhaServico{{"immich-app", saidaEnvAusente(dir)}},
		map[string]string{"immich-app": compose},
		func(string) (string, error) {
			t.Error("não deveria repetir quando o usuário recusa")
			return "", nil
		})

	if len(restantes) != 1 {
		t.Errorf("a falha deveria continuar: %+v", restantes)
	}
	if existe(filepath.Join(dir, ".env")) {
		t.Error("o .env não deveria ser criado quando o usuário recusa")
	}
}
