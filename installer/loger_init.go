package installer

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
)

//go:embed loger/loger.service
var logerUnit string

//go:embed loger/loger.openrc
var logerOpenRC string

//go:embed loger/loger.runit
var logerRunit string

const (
	LOGER_UNIT   = "/etc/systemd/system/loger.service"
	LOGER_OPENRC = "/etc/init.d/loger"
)

type initLoger struct {
	nome        string
	arquivo     string
	conteudo    string
	modo        os.FileMode
	link        string
	ativar      [][]string
	desativar   [][]string
	aposRemover [][]string
	status      string
	logs        string
}

func detectarInitLoger(dir func(string) bool, bin func(string) bool) *initLoger {
	switch {
	case dir("/run/systemd/system") && bin("systemctl"):
		return &initLoger{
			nome:        "systemd",
			arquivo:     LOGER_UNIT,
			conteudo:    logerUnit,
			modo:        0o644,
			ativar:      [][]string{{"systemctl", "daemon-reload"}, {"systemctl", "enable", "--now", "loger"}},
			desativar:   [][]string{{"systemctl", "disable", "--now", "loger"}},
			aposRemover: [][]string{{"systemctl", "daemon-reload"}},
			status:      "systemctl status loger",
			logs:        "journalctl -u loger -n 50",
		}
	case dir("/run/openrc") && bin("rc-update") && bin("rc-service"):
		return &initLoger{
			nome:      "OpenRC",
			arquivo:   LOGER_OPENRC,
			conteudo:  logerOpenRC,
			modo:      0o755,
			ativar:    [][]string{{"rc-update", "add", "loger", "default"}, {"rc-service", "loger", "start"}},
			desativar: [][]string{{"rc-service", "loger", "stop"}, {"rc-update", "del", "loger", "default"}},
			status:    "rc-service loger status",
			logs:      "rc-service loger status",
		}
	case bin("sv") && bin("runsvdir"):
		base, servicos := pastasRunit(dir)
		if servicos == "" {
			return nil
		}
		return &initLoger{
			nome:      "runit",
			arquivo:   filepath.Join(base, "loger", "run"),
			conteudo:  logerRunit,
			modo:      0o755,
			link:      filepath.Join(servicos, "loger"),
			desativar: [][]string{{"sv", "down", "loger"}},
			status:    "sv status loger",
			logs:      "sv status loger",
		}
	}
	return nil
}

func pastasRunit(dir func(string) bool) (base, servicos string) {
	if dir("/etc/runit/sv") && dir("/run/runit/service") {
		return "/etc/runit/sv", "/run/runit/service"
	}
	for _, s := range []string{"/var/service", "/etc/service"} {
		if dir(s) {
			return "/etc/sv", s
		}
	}
	return "", ""
}

func arquivosServicoLoger() []string {
	return []string{
		"/var/service/loger",
		"/etc/service/loger",
		"/run/runit/service/loger",
		"/etc/sv/loger",
		"/etc/runit/sv/loger",
		LOGER_OPENRC,
		LOGER_UNIT,
	}
}

func existeLink(caminho string) bool {
	_, err := os.Lstat(caminho)
	return err == nil
}

func (i *initLoger) instalado() bool {
	if i.link != "" {
		return existe(i.arquivo) && existeLink(i.link)
	}
	return existe(i.arquivo)
}

func (i *initLoger) criar() error {
	if err := os.MkdirAll(filepath.Dir(i.arquivo), 0o755); err != nil {
		return err
	}
	if !existe(i.arquivo) {
		if err := os.WriteFile(i.arquivo, []byte(i.conteudo), i.modo); err != nil {
			return err
		}
	}
	if i.link != "" && !existeLink(i.link) {
		return os.Symlink(filepath.Dir(i.arquivo), i.link)
	}
	return nil
}

func rodarComandos(cmds [][]string) (string, error) {
	for _, c := range cmds {
		if _, stderr, err := run(c[0], c[1:]...); err != nil {
			return stderr, err
		}
	}
	return "", nil
}

func juntarComandos(cmds [][]string) string {
	partes := make([]string, len(cmds))
	for i, c := range cmds {
		partes[i] = strings.Join(c, " ")
	}
	return strings.Join(partes, " && ")
}
