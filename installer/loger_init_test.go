package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ambiente(dirs, bins []string) (func(string) bool, func(string) bool) {
	tem := func(lista []string) func(string) bool {
		return func(s string) bool {
			for _, x := range lista {
				if x == s {
					return true
				}
			}
			return false
		}
	}
	return tem(dirs), tem(bins)
}

func TestDetectarInitLoger(t *testing.T) {
	casos := []struct {
		nome    string
		dirs    []string
		bins    []string
		init    string
		arquivo string
		link    string
	}{
		{"systemd", []string{"/run/systemd/system"}, []string{"systemctl"}, "systemd", LOGER_UNIT, ""},
		{"openrc", []string{"/run/openrc"}, []string{"rc-update", "rc-service"}, "OpenRC", LOGER_OPENRC, ""},
		{"runit void", []string{"/var/service"}, []string{"sv", "runsvdir"}, "runit", "/etc/sv/loger/run", "/var/service/loger"},
		{"runit etc/service", []string{"/etc/service"}, []string{"sv", "runsvdir"}, "runit", "/etc/sv/loger/run", "/etc/service/loger"},
		{"runit artix", []string{"/etc/runit/sv", "/run/runit/service", "/var/service"}, []string{"sv", "runsvdir"}, "runit", "/etc/runit/sv/loger/run", "/run/runit/service/loger"},
		{"systemd ganha de openrc", []string{"/run/systemd/system", "/run/openrc"}, []string{"systemctl", "rc-update", "rc-service"}, "systemd", LOGER_UNIT, ""},
		{"systemctl sem systemd rodando", []string{}, []string{"systemctl"}, "", "", ""},
		{"openrc instalado mas não em uso", []string{}, []string{"rc-update", "rc-service"}, "", "", ""},
		{"runit sem pasta de serviços", []string{}, []string{"sv", "runsvdir"}, "", "", ""},
		{"nenhum", []string{}, []string{}, "", "", ""},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			dir, bin := ambiente(c.dirs, c.bins)
			i := detectarInitLoger(dir, bin)
			if c.init == "" {
				if i != nil {
					t.Fatalf("esperava nenhum init, veio %s", i.nome)
				}
				return
			}
			if i == nil {
				t.Fatalf("esperava %s, veio nil", c.init)
			}
			if i.nome != c.init || i.arquivo != c.arquivo || i.link != c.link {
				t.Fatalf("veio %s %s %s", i.nome, i.arquivo, i.link)
			}
			if i.status == "" || i.logs == "" {
				t.Fatal("status e logs não podem ser vazios")
			}
		})
	}
}

func TestArquivosDeServicoEmbutidos(t *testing.T) {
	if !strings.Contains(logerUnit, "ExecStart=/usr/local/bin/loger daemon") {
		t.Fatal("unit do systemd")
	}
	if !strings.HasPrefix(logerOpenRC, "#!/sbin/openrc-run") || !strings.Contains(logerOpenRC, `command="`+LOGER_BIN+`"`) {
		t.Fatal("script do OpenRC")
	}
	if !strings.HasPrefix(logerRunit, "#!/bin/sh") || !strings.Contains(logerRunit, "exec "+LOGER_BIN+" daemon") {
		t.Fatal("script do runit")
	}
}

func TestCriarRunitComLink(t *testing.T) {
	tmp := t.TempDir()
	servicos := filepath.Join(tmp, "var", "service")
	if err := os.MkdirAll(servicos, 0o755); err != nil {
		t.Fatal(err)
	}
	i := &initLoger{
		arquivo:  filepath.Join(tmp, "etc", "sv", "loger", "run"),
		conteudo: logerRunit,
		modo:     0o755,
		link:     filepath.Join(servicos, "loger"),
	}
	if i.instalado() {
		t.Fatal("não deveria estar instalado")
	}
	if err := i.criar(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(i.arquivo)
	if err != nil || info.Mode().Perm() != 0o755 {
		t.Fatalf("script run: %v %v", err, info)
	}
	alvo, err := os.Readlink(i.link)
	if err != nil || alvo != filepath.Dir(i.arquivo) {
		t.Fatalf("link: %v %s", err, alvo)
	}
	if !i.instalado() {
		t.Fatal("deveria estar instalado")
	}
	if err := i.criar(); err != nil {
		t.Fatalf("criar de novo deveria ser idempotente: %v", err)
	}
}

func TestCriarNaoSobrescreve(t *testing.T) {
	tmp := t.TempDir()
	i := &initLoger{arquivo: filepath.Join(tmp, "init.d", "loger"), conteudo: logerOpenRC, modo: 0o755}
	if err := os.MkdirAll(filepath.Dir(i.arquivo), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(i.arquivo, []byte("ajustado à mão"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := i.criar(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(i.arquivo)
	if string(b) != "ajustado à mão" {
		t.Fatal("sobrescreveu um serviço existente")
	}
}

func TestJuntarComandos(t *testing.T) {
	got := juntarComandos([][]string{{"rc-update", "add", "loger", "default"}, {"rc-service", "loger", "start"}})
	if got != "rc-update add loger default && rc-service loger start" {
		t.Fatal(got)
	}
}
