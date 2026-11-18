package installer

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

var opcoesMenu = map[string]string{"1": "install", "2": "uninstall", "3": "media disk", "4": "language", "5": "exit"}

func interromperInstalacao(motivo string) {
	PrintText(fmt.Sprintf("\n%s\n\n", ColorText(Tr("Instalação interrompida.", "Installation aborted."), RED+BOLD)))
	slog.Warn("Installation aborted", "reason", motivo)
	Registrar("install", "result", "aborted", "reason", motivo)
}

func etapasInstalacao() (map[string]string, error) {
	if err := CheckOS(); err != nil {
		return nil, err
	}

	engineChecks := []struct {
		fn    func() error
		label string
	}{
		{checkEngineBinary, CONTAINER_ENGINE},
		{checkComposePlugin, CONTAINER_ENGINE + " compose"},
		{checkEngineDaemon, CONTAINER_ENGINE + " daemon"},
	}
	for _, c := range engineChecks {
		if !checkOptional(c.fn, c.label) {
			interromperInstalacao(fmt.Sprintf("%s check failed", c.label))
			return nil, nil
		}
	}

	if err := setupEnv(); err != nil {
		return nil, err
	}

	if !resolveMediaMountInteractive() {
		interromperInstalacao(fmt.Sprintf("user chose not to continue without %s mounted", MEDIA_DIR))
		return nil, nil
	}

	if err := checkDiskSpace(); err != nil {
		return nil, err
	}

	todosCompose := findComposeFiles()
	if len(todosCompose) == 0 {
		return nil, errors.New(Tr("Nenhum arquivo docker-compose.yml/yaml encontrado nas pastas do repositório.",
			"No docker-compose.yml/yaml file found in the repository folders."))
	}
	composeFiles := escolherServicos(todosCompose)
	Registrar("services_selected", "services", nomesOrdenados(composeFiles))
	ativarLoger := perguntarLoger()
	if err := checkPorts(composeFiles); err != nil {
		return nil, err
	}

	if err := createDirectories(); err != nil {
		return nil, err
	}
	configurePermissions()

	if err := validateCompose(composeFiles); err != nil {
		return nil, err
	}
	if err := startServices(composeFiles); err != nil {
		return nil, err
	}
	if ativarLoger {
		instalarLoger()
	}
	return composeFiles, nil
}

func fluxoInstalar() {
	PrintHeader(Tr("INSTALAÇÃO", "INSTALLATION"))
	if DRY_RUN {
		PrintWarning(Tr("Modo --dry-run ativo: nenhuma alteração real será feita.", "--dry-run mode on: no real changes will be made."))
	}

	CONTAINER_ENGINE = chooseEngine()
	Registrar("install", "result", "started", "engine", CONTAINER_ENGINE)

	composeFiles, err := etapasInstalacao()
	if err != nil {
		PrintFail(err.Error())
		interromperInstalacao(err.Error())
		return
	}
	if composeFiles == nil {
		return
	}

	Registrar("install", "result", "completed", "services", nomesOrdenados(composeFiles))
	PrintHeader(Tr("INSTALAÇÃO CONCLUÍDA", "INSTALLATION COMPLETE"))
	slog.Info("Installation completed")
}

func fluxoDesinstalar() {
	PrintHeader(Tr("DESINSTALAÇÃO", "UNINSTALLATION"))

	LoadEnvIntoState()
	CONTAINER_ENGINE = chooseEngine()
	Registrar("uninstall", "result", "started", "engine", CONTAINER_ENGINE)

	composeFiles := findComposeFiles()
	stopServices(composeFiles)
	Registrar("containers_stopped", "services", nomesOrdenados(composeFiles))
	desinstalarLoger()

	PrintInfo(fmt.Sprintf(Tr("\n%s NUNCA é tocado por esta opção — seus arquivos de mídia estão seguros.",
		"\n%s is NEVER touched by this option — your media files are safe."), MEDIA_DIR))

	alvos := collectDataDirs(composeFiles)
	if len(alvos) == 0 {
		PrintInfo(Tr("Nenhuma pasta de dados de serviço encontrada no disco.", "No service data folders found on disk."))
		PrintInfo(Tr("Apenas os containers foram parados.", "Only the containers were stopped."))
		Registrar("uninstall", "result", "completed", "configs", "no folders found")
		return
	}

	PrintSection(Tr("Pastas de dados encontradas no disco:", "Data folders found on disk:"))
	for _, alvo := range alvos {
		PrintInfo(fmt.Sprintf("%s  (%s)", alvo, tamanhoLegivel(alvo)))
	}

	resposta := readLine(Tr("\nTambém remover as pastas listadas acima? "+
		"Isso apaga bancos de dados e configs dos serviços. [s/N]: ",
		"\nAlso remove the folders listed above? "+
			"This deletes the services' databases and configs. [y/N]: "))
	if !Confirmou(resposta, false) {
		PrintInfo(Tr("Configurações mantidas. Apenas os containers foram parados.", "Configs kept. Only the containers were stopped."))
		Registrar("uninstall", "result", "completed", "configs", "kept")
		return
	}

	aviso := fmt.Sprintf(Tr("Isso vai apagar PERMANENTEMENTE as pastas listadas acima: configs, "+
		"bancos de dados (File Browser, Navidrome, etc.) e caches.\n"+
		"    %s continua intocado.",
		"This will PERMANENTLY delete the folders listed above: configs, "+
			"databases (File Browser, Navidrome, etc.) and caches.\n"+
			"    %s stays untouched."), MEDIA_DIR)
	if !confirmDangerous(aviso, PalavraConfirmacao()) {
		PrintInfo(Tr("Operação cancelada — configurações mantidas.", "Operation cancelled — configs kept."))
		Registrar("uninstall", "result", "completed", "configs", "kept (confirmation cancelled)")
		return
	}

	if DRY_RUN {
		for _, alvo := range alvos {
			PrintInfo(Tr("[dry-run] removeria ", "[dry-run] would remove ") + alvo)
		}
		Registrar("uninstall", "result", "completed",
			"configs", fmt.Sprintf("[dry-run] would remove %d folder(s)", len(alvos)))
		return
	}

	removidas := []string{}
	for _, alvo := range alvos {
		slog.Warn("Removing", "path", alvo)
		os.RemoveAll(alvo)
		if existe(alvo) {
			PrintWarning(alvo + Tr(" não pôde ser removido por completo", " could not be fully removed"))
		} else {
			PrintOk(alvo + Tr(" removido", " removed"))
			removidas = append(removidas, alvo)
		}
	}
	Registrar("uninstall", "result", "completed", "configs_removed", removidas)
}

func MainMenu() {
	LoadEnvIntoState()
	for {
		PrintHeader(Tr("INSTALADOR DO MEDIA SERVER", "MEDIA SERVER INSTALLER"))
		if DRY_RUN {
			PrintWarning(Tr("Modo --dry-run ativo\n", "--dry-run mode on\n"))
		}

		PrintInfo(Tr("1) Instalar / atualizar serviços", "1) Install / update services"))
		PrintInfo(Tr("2) Desinstalar serviços", "2) Uninstall services"))
		PrintInfo(fmt.Sprintf(Tr("3) Gerenciar disco de mídia (%s)", "3) Manage media disk (%s)"), MEDIA_DIR))
		PrintInfo("4) Idioma / Language")
		PrintInfo(Tr("5) Sair", "5) Exit"))

		escolha := readLine(Tr("\nEscolha uma opção: ", "\nChoose an option: "))
		slog.Info("Option chosen", "option", escolha)

		significado, ok := opcoesMenu[escolha]
		if !ok {
			significado = "invalid"
		}
		Registrar("menu", "option", escolha, "meaning", significado)

		switch escolha {
		case "1":
			fluxoInstalar()
			pressEnterToContinue()
		case "2":
			fluxoDesinstalar()
			pressEnterToContinue()
		case "3":
			MediaMountMenu()
		case "4":
			EscolherIdioma()
		case "5":
			PrintInfo(Tr("Até mais!", "Goodbye!"))
			return
		default:
			PrintFail(Tr("Opção inválida.", "Invalid option."))
		}
	}
}

const (
	MODO_CLI = "cli"
	MODO_WEB = "web"
)

// EscolherModo pergunta se o usuário quer seguir pelo terminal ou pela
// interface web. Enter sem nada escolhe o terminal.
func EscolherModo() string {
	for {
		PrintHeader(Tr("MODO DE INSTALAÇÃO", "INSTALLATION MODE"))
		PrintInfo(Tr("1) Terminal (CLI) (padrão)", "1) Terminal (CLI) (default)"))
		PrintInfo(Tr("2) Navegador (Web)", "2) Browser (Web)"))

		escolha := readLine(Tr("\nComo deseja usar o instalador? [1]: ", "\nHow do you want to use the installer? [1]: "))
		switch escolha {
		case "", "1":
			slog.Info("Mode chosen", "mode", MODO_CLI)
			Registrar("mode", "mode", MODO_CLI)
			return MODO_CLI
		case "2":
			slog.Info("Mode chosen", "mode", MODO_WEB)
			Registrar("mode", "mode", MODO_WEB)
			return MODO_WEB
		default:
			PrintFail(Tr("Opção inválida.", "Invalid option."))
		}
	}
}
