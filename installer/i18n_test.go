package installer

import "testing"

func TestNormalizarIdioma(t *testing.T) {
	casos := map[string]string{
		"pt": IDIOMA_PT, "PT": IDIOMA_PT, "pt-BR": IDIOMA_PT, "pt_BR.UTF-8": IDIOMA_PT,
		"en": IDIOMA_EN, "en_US.UTF-8": IDIOMA_EN, "en-GB": IDIOMA_EN,
		"C": "", "POSIX": "", "": "", "es_ES.UTF-8": "", "portugues": "", "english": "",
	}
	for entrada, esperado := range casos {
		if obtido := NormalizarIdioma(entrada); obtido != esperado {
			t.Errorf("NormalizarIdioma(%q) = %q, esperado %q", entrada, obtido, esperado)
		}
	}
}

func TestIdiomaDoAmbiente(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "en_US.UTF-8")
	if got := IdiomaDoAmbiente(); got != IDIOMA_EN {
		t.Errorf("LANG=en_US: obtido %q", got)
	}
	t.Setenv("LC_ALL", "pt_BR.UTF-8")
	if got := IdiomaDoAmbiente(); got != IDIOMA_PT {
		t.Errorf("LC_ALL tem prioridade sobre LANG: obtido %q", got)
	}
	t.Setenv("LC_ALL", "C")
	if got := IdiomaDoAmbiente(); got != "" {
		t.Errorf("LC_ALL=C deveria ser indefinido: obtido %q", got)
	}
}

func TestConfirmou(t *testing.T) {
	for _, r := range []string{"s", "S", "sim", "y", "Y", "yes", " yes "} {
		if !Confirmou(r, false) {
			t.Errorf("Confirmou(%q, false) deveria ser true", r)
		}
	}
	for _, r := range []string{"n", "N", "nao", "não", "no"} {
		if Confirmou(r, true) {
			t.Errorf("Confirmou(%q, true) deveria ser false", r)
		}
	}
	for _, r := range []string{"", "talvez", "ok", "confirmo"} {
		if Confirmou(r, false) {
			t.Errorf("Confirmou(%q, false) deveria cair no padrão false", r)
		}
		if !Confirmou(r, true) {
			t.Errorf("Confirmou(%q, true) deveria cair no padrão true", r)
		}
	}
}

func TestTrEPalavraConfirmacao(t *testing.T) {
	antes := IDIOMA
	defer func() { IDIOMA = antes }()

	IDIOMA = IDIOMA_PT
	if Tr("sim", "yes") != "sim" || PalavraConfirmacao() != "CONFIRMO" {
		t.Error("pt deveria devolver o texto em português")
	}
	IDIOMA = IDIOMA_EN
	if Tr("sim", "yes") != "yes" || PalavraConfirmacao() != "CONFIRM" {
		t.Error("en deveria devolver o texto em inglês")
	}
}
