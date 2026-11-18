package installer

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	IDIOMA_PT = "pt"
	IDIOMA_EN = "en"
)

var IDIOMA = IDIOMA_PT

func Tr(pt, en string) string {
	if IDIOMA == IDIOMA_EN {
		return en
	}
	return pt
}

func NormalizarIdioma(valor string) string {
	valor = strings.ToLower(strings.TrimSpace(valor))
	switch {
	case valor == "pt" || strings.HasPrefix(valor, "pt-") || strings.HasPrefix(valor, "pt_"):
		return IDIOMA_PT
	case valor == "en" || strings.HasPrefix(valor, "en-") || strings.HasPrefix(valor, "en_"):
		return IDIOMA_EN
	}
	return ""
}

func IdiomaDoAmbiente() string {
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if valor := os.Getenv(v); valor != "" {
			return NormalizarIdioma(valor)
		}
	}
	return ""
}

var aoTrocarIdioma func()

func EscolherIdioma() {
	padrao := IDIOMA
	if idioma := IdiomaDoAmbiente(); idioma != "" && atual == nil {
		padrao = idioma
	}
	numeroPadrao := "1"
	if padrao == IDIOMA_EN {
		numeroPadrao = "2"
	}
	for {
		PrintHeader("IDIOMA / LANGUAGE")
		PrintInfo("1) Português")
		PrintInfo("2) English")
		escolha := readLine(fmt.Sprintf("\nIdioma / Language [%s]: ", numeroPadrao))
		if escolha == "" {
			escolha = numeroPadrao
		}
		switch escolha {
		case "1":
			definirIdioma(IDIOMA_PT)
			return
		case "2":
			definirIdioma(IDIOMA_EN)
			return
		}
		PrintFail("Opção inválida. / Invalid option.")
	}
}

func definirIdioma(idioma string) {
	IDIOMA = idioma
	if atual != nil {
		slog.Info("Language changed", "language", idioma)
		Registrar("language", "language", idioma)
	}
	if aoTrocarIdioma != nil {
		aoTrocarIdioma()
	}
}

func Confirmou(resposta string, padrao bool) bool {
	switch strings.ToLower(strings.TrimSpace(resposta)) {
	case "s", "sim", "y", "yes":
		return true
	case "n", "nao", "não", "no":
		return false
	}
	return padrao
}

func PalavraConfirmacao() string {
	return Tr("CONFIRMO", "CONFIRM")
}
