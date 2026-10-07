package main

import (
	"fmt"
	_ "log"
	"os"
	_ "strings"

	"self-hosted-media-server/installer"
)

const ADVERTENCIA = "\033[1;31mADVERTÊNCIA:\033[0m Este script é fornecido \"como está\" e não se responsabiliza por quaisquer danos ou perda de dados. Use por sua própria conta e risco."
const ASCII_ART = ` ___           _        _           _              ____       _  __       _   _           _           _
|_ _|_ __  ___| |_ __ _| | __ _  __| | ___  _ __  / ___|  ___| |/ _|     | | | | ___  ___| |_ ___  __| |
 | || '_ \/ __| __/ _` + "`" + ` | |/ _` + "`" + ` |/ _` + "`" + ` |/ _ \| '__| \___ \ / _ \ | |_ _____| |_| |/ _ \/ __| __/ _ \/ _` + "`" + ` |
 | || | | \__ \ || (_| | | (_| | (_| | (_) | |     ___) |  __/ |  _|_____|  _  | (_) \__ \ ||  __/ (_| |
|___|_| |_|___/\__\__,_|_|\__,_|\__,_|\___/|_|    |____/ \___|_|_|       |_| |_|\___/|___/\__\___|\__,_|`

func main() {
	mostrarAviso()

	err := installer.CheckOS()
	if err != nil {
		fmt.Println("Erro:", err)
		os.Exit(1)
	}

	err = installer.CheckPrivileges()
	if err != nil {
		fmt.Println("Erro:", err)
		os.Exit(1)
	}
}

func mostrarAviso() {
	fmt.Println(ASCII_ART)
	fmt.Println("Instalador iniciado")
	fmt.Println(ADVERTENCIA)
}
