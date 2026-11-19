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

func advertencia() string {
	return installer.Tr(
		"\033[1;31mADVERTÊNCIA:\033[0m Este script é fornecido \"como está\" e não se responsabiliza por quaisquer danos ou perda de dados. Use por sua própria conta e risco.",
		"\033[1;31mWARNING:\033[0m This script is provided \"as is\" with no liability for any damage or data loss. Use at your own risk.")
}

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
		"Simulate every action without changing anything / Simula todas as ações sem executar nada de fato.")
	portaWeb := flag.Int("web-porta", 8787, "Web UI port (browser mode); 0 picks a free port / Porta da interface web; 0 escolhe uma porta livre.")
	mostrarVersao := flag.Bool("version", false, "Print the installer version and exit / Mostra a versão do instalador e sai.")
	flag.Parse()

	if *mostrarVersao {
		fmt.Println(versao)
		return 0
	}

	fmt.Print(ASCII_ART + "\n")
	installer.EscolherIdioma()
	fmt.Println(advertencia())

	if err := installer.CheckPrivileges(); err != nil {
		installer.PrintFail(err.Error())
		return 1
	}

	if err := installer.SetupLogging(installer.LOG_FILE); err != nil {
		installer.PrintFail(fmt.Sprintf(installer.Tr("Não foi possível abrir o log %s: %v", "Could not open the log %s: %v"), installer.LOG_FILE, err))
		return 1
	}
	slog.Info("Installer started", "version", versao, "dry_run", installer.DRY_RUN, "language", installer.IDIOMA)
	installer.IniciarExecucao()

	sinais := make(chan os.Signal, 1)
	signal.Notify(sinais, os.Interrupt)
	go func() {
		<-sinais
		fmt.Print("\n\n")
		installer.PrintWarning(installer.Tr("Encerrado pelo usuário.", "Stopped by the user."))
		slog.Warn("Stopped by the user (Ctrl+C)")
		installer.FinalizarExecucao("cancelled (Ctrl+C)")
		os.Exit(130)
	}()

	defer func() {
		if r := recover(); r != nil {
			slog.Error("Unexpected error", "error", fmt.Sprint(r))
			installer.Registrar("unexpected_error", "error", fmt.Sprint(r))
			installer.FinalizarExecucao("error")
			installer.PrintFail(installer.Tr("Erro inesperado. Detalhes em ", "Unexpected error. Details in ") + installer.LOG_FILE)
			codigo = 1
		}
	}()

	switch installer.EscolherModo() {
	case installer.MODO_WEB:
		if err := installer.ModoWeb(*portaWeb); err != nil {
			installer.PrintFail(err.Error())
			installer.FinalizarExecucao("error")
			return 1
		}
	default:
		installer.MainMenu()
	}

	slog.Info("Installer finished")
	installer.FinalizarExecucao("finished")
	installer.PrintInfo(installer.Tr("Log desta execução: ", "Log of this run: ") + installer.LOG_FILE)
	installer.PrintInfo(installer.Tr("Histórico: ", "History: ") + installer.HistoryFile)
	return 0
}
