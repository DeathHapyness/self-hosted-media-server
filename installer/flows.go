package installer

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var opcoesMenu = map[string]string{"1": "instalar", "2": "desinstalar", "3": "disco de mídia", "4": "sair"}

func interromperInstalacao(motivo string) {
	fmt.Printf("\n%s\n\n", ColorText("Instalação interrompida.", RED+BOLD))
	slog.Warn("Instalação interrompida", "motivo", motivo)
	Registrar("instalacao", "resultado", "interrompida", "motivo", motivo)
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
			interromperInstalacao(fmt.Sprintf("verificação de %s falhou", c.label))
			return nil, nil
		}
	}

	if err := setupEnv(); err != nil {
		return nil, err
	}

	if !resolveMediaMountInteractive() {
		interromperInstalacao(fmt.Sprintf("usuário não quis continuar sem %s montado", MEDIA_DIR))
		return nil, nil
	}

	if err := checkDiskSpace(); err != nil {
		return nil, err
	}

	todosCompose := findComposeFiles()
	if len(todosCompose) == 0 {
		return nil, errors.New("Nenhum arquivo docker-compose.yml/yaml encontrado nas pastas do repositório.")
	}
	composeFiles := escolherServicos(todosCompose)
	Registrar("servicos_escolhidos", "servicos", nomesOrdenados(composeFiles))
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
	return composeFiles, nil
}

func fluxoInstalar() {
	PrintHeader("INSTALAÇÃO")
	if DRY_RUN {
		PrintWarning("Modo --dry-run ativo: nenhuma alteração real será feita.")
	}

	CONTAINER_ENGINE = chooseEngine()
	Registrar("instalacao", "resultado", "iniciada", "engine", CONTAINER_ENGINE)

	composeFiles, err := etapasInstalacao()
	if err != nil {
		PrintFail(err.Error())
		interromperInstalacao(err.Error())
		return
	}
	if composeFiles == nil {
		return
	}

	Registrar("instalacao", "resultado", "concluida", "servicos", nomesOrdenados(composeFiles))
	PrintHeader("INSTALLATION COMPLETE")
	slog.Info("Instalação concluída")
}

func fluxoDesinstalar() {
	PrintHeader("DESINSTALAÇÃO")

	LoadEnvIntoState()
	CONTAINER_ENGINE = chooseEngine()
	Registrar("desinstalacao", "resultado", "iniciada", "engine", CONTAINER_ENGINE)

	composeFiles := findComposeFiles()
	stopServices(composeFiles)
	Registrar("containers_parados", "servicos", nomesOrdenados(composeFiles))

	PrintInfo(fmt.Sprintf("\n%s NUNCA é tocado por esta opção — seus arquivos de mídia estão seguros.", MEDIA_DIR))

	alvos := collectDataDirs(composeFiles)
	if len(alvos) == 0 {
		PrintInfo("Nenhuma pasta de dados de serviço encontrada no disco.")
		PrintInfo("Apenas os containers foram parados.")
		Registrar("desinstalacao", "resultado", "concluida", "configs", "nenhuma pasta encontrada")
		return
	}

	PrintSection("Pastas de dados encontradas no disco:")
	for _, alvo := range alvos {
		PrintInfo(fmt.Sprintf("%s  (%s)", alvo, tamanhoLegivel(alvo)))
	}

	resposta := strings.ToLower(readLine("\nTambém remover as pastas listadas acima? " +
		"Isso apaga bancos de dados e configs dos serviços. [s/N]: "))
	if resposta != "s" {
		PrintInfo("Configurações mantidas. Apenas os containers foram parados.")
		Registrar("desinstalacao", "resultado", "concluida", "configs", "mantidas")
		return
	}

	aviso := fmt.Sprintf("Isso vai apagar PERMANENTEMENTE as pastas listadas acima: configs, "+
		"bancos de dados (File Browser, Navidrome, etc.) e caches.\n"+
		"    %s continua intocado.", MEDIA_DIR)
	if !confirmDangerous(aviso, "CONFIRMO") {
		PrintInfo("Operação cancelada — configurações mantidas.")
		Registrar("desinstalacao", "resultado", "concluida", "configs", "mantidas (confirmação cancelada)")
		return
	}

	if DRY_RUN {
		for _, alvo := range alvos {
			PrintInfo("[dry-run] removeria " + alvo)
		}
		Registrar("desinstalacao", "resultado", "concluida",
			"configs", fmt.Sprintf("[dry-run] removeria %d pasta(s)", len(alvos)))
		return
	}

	removidas := []string{}
	for _, alvo := range alvos {
		slog.Warn("Removendo", "caminho", alvo)
		os.RemoveAll(alvo)
		if existe(alvo) {
			PrintWarning(alvo + " não pôde ser removido por completo")
		} else {
			PrintOk(alvo + " removido")
			removidas = append(removidas, alvo)
		}
	}
	Registrar("desinstalacao", "resultado", "concluida", "configs_removidas", removidas)
}

func MainMenu() {
	LoadEnvIntoState()
	for {
		PrintHeader("MEDIA SERVER INSTALLER")
		if DRY_RUN {
			PrintWarning("Modo --dry-run ativo\n")
		}

		PrintInfo("1) Instalar / atualizar serviços")
		PrintInfo("2) Desinstalar serviços")
		PrintInfo(fmt.Sprintf("3) Gerenciar disco de mídia (%s)", MEDIA_DIR))
		PrintInfo("4) Sair")

		escolha := readLine("\nEscolha uma opção: ")
		slog.Info("Opção escolhida", "opcao", escolha)

		significado, ok := opcoesMenu[escolha]
		if !ok {
			significado = "inválida"
		}
		Registrar("menu", "opcao", escolha, "significado", significado)

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
			PrintInfo("Até mais!")
			return
		default:
			PrintFail("Opção inválida.")
		}
	}
}
