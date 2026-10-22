package installer

import (
	"log/slog"
	"os"
)

type Interface interface {
	Cabecalho(texto string)
	Secao(texto string)
	Ok(texto string)
	Falha(texto string)
	Aviso(texto string)
	Info(texto string)
	Perigo(texto string)
	Texto(texto string)
	Perguntar(prompt string) (string, error)
}

var UI Interface = Terminal{}

func PrintHeader(text string) {
	UI.Cabecalho(text)
}

func PrintSection(text string) {
	UI.Secao(text)
}

func PrintOk(text string) {
	UI.Ok(text)
}

func PrintFail(text string) {
	UI.Falha(text)
	slog.Error(text)
}

func PrintWarning(text string) {
	UI.Aviso(text)
}

func PrintInfo(text string) {
	UI.Info(text)
}

func PrintDanger(text string) {
	UI.Perigo(text)
}

func PrintText(text string) {
	UI.Texto(text)
}

func readLine(prompt string) string {
	linha, err := UI.Perguntar(prompt)
	if err != nil {
		slog.Warn("Entrada encerrada; saindo", "erro", err)
		FinalizarExecucao("cancelada (entrada encerrada)")
		os.Exit(1)
	}
	return linha
}
