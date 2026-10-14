package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"self-hosted-media-server/installer"
)

var versao = "dev"

const ADVERTENCIA = "\033[1;31mADVERTÊNCIA:\033[0m Este script é fornecido \"como está\" e não se responsabiliza por quaisquer danos ou perda de dados. Use por sua própria conta e risco."
const ASCII_ART = `  ___           _        _           _              ____       _  __       _   _           _           _
 |_ _|_ __  ___| |_ __ _| | __ _  __| | ___  _ __  / ___|  ___| |/ _|     | | | | ___  ___| |_ ___  __| |
 | || '_ \/ __| __/ _` + "`" + ` | |/ _` + "`" + ` |/ _` + "`" + ` |/ _ \| '__| \___ \ / _ \ | |_ _____| |_| |/ _ \/ __| __/ _ \/ _` + "`" + ` |
 | || | | \__ \ || (_| | | (_| | (_| | (_) | |     ___) |  __/ |  _|_____|  _  | (_) \__ \ ||  __/ (_| |
|___|_| |_|___/\__\__,_|_|\__,_|\__,_|\___/|_|    |____/ \___|_|_|       |_| |_|\___/|___/\__\___|\__,_|
`

func main() {
	os.Exit(run())
}

func run() (codigo int) {
	flag.BoolVar(&installer.DRY_RUN, "dry-run", false,
		"Simula todas as ações sem executar nada de fato (nenhum arquivo criado, nenhum comando destrutivo rodado).")
	mostrarVersao := flag.Bool("version", false, "Mostra a versão do instalador e sai.")
	flag.Parse()

	if *mostrarVersao {
		fmt.Println(versao)
		return 0
	}

	fmt.Print(ASCII_ART + "\n")
	fmt.Println(ADVERTENCIA)

	if err := installer.CheckPrivileges(); err != nil {
		installer.PrintFail(err.Error())
		return 1
	}

	if err := installer.SetupLogging(installer.LOG_FILE); err != nil {
		installer.PrintFail(fmt.Sprintf("Não foi possível abrir o log %s: %v", installer.LOG_FILE, err))
		return 1
	}
	slog.Info("Instalador iniciado", "versao", versao, "dry_run", installer.DRY_RUN)
	installer.IniciarExecucao()

	sinais := make(chan os.Signal, 1)
	signal.Notify(sinais, os.Interrupt)
	go func() {
		<-sinais
		fmt.Print("\n\n")
		installer.PrintWarning("Encerrado pelo usuário.")
		slog.Warn("Encerrado pelo usuário (Ctrl+C)")
		installer.FinalizarExecucao("cancelada (Ctrl+C)")
		os.Exit(130)
	}()

	defer func() {
		if r := recover(); r != nil {
			slog.Error("Erro inesperado", "erro", fmt.Sprint(r))
			installer.Registrar("erro_inesperado", "erro", fmt.Sprint(r))
			installer.FinalizarExecucao("erro")
			installer.PrintFail("Erro inesperado. Detalhes em " + installer.LOG_FILE)
			codigo = 1
		}
	}()

	installer.MainMenu()

	slog.Info("Instalador finalizado")
	installer.FinalizarExecucao("finalizada")
	installer.PrintInfo("Log desta execução: " + installer.LOG_FILE)
	installer.PrintInfo("Histórico: " + installer.HistoryFile)
	return 0
}
