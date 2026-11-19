# 🛡️ Loger

Auditoria leve para o servidor: registra logins (SSH e locais), comandos
executados (inclusive depois de `sudo`/`su`), uso de privilégios, alterações em
arquivos sensíveis (`/etc/passwd`, `sudoers`, `authorized_keys`, cron,
systemd…), contas criadas/alteradas, portas novas e conexões suspeitas.

Projeto: [DeathHapyness/Loger](https://github.com/DeathHapyness/Loger)

Diferente dos outros serviços deste repositório, o Loger **não roda em
container**: é um binário estático que roda direto no host, como serviço do
init do sistema (systemd, OpenRC ou runit). Ele não abre nenhuma porta.

## Ativado por padrão

O Loger **vem ativado** no instalador. Ele não aparece na lista numerada de
serviços; logo depois dela, o instalador pergunta:

```text
Manter o Loger ativado? [S/n]:
```

- `Enter` ou `s`: instala o Loger depois de subir os outros serviços.
- `n`: não instala o Loger nesta execução.

No modo web, a mesma pergunta aparece com os botões **Sim** / **Não**.

## O que o instalador faz

1. Baixa a imagem `ghcr.io/deathhapyness/loger:latest` e copia o binário dela
   para `/usr/local/bin/loger`. O container é só criado e removido em seguida,
   nunca é iniciado. Isso já entrega o binário da arquitetura certa
   (amd64 ou arm64).
2. Detecta o init do sistema e cria o serviço dele:

| Init | Distros (exemplos) | Arquivo criado | Como ativa |
|---|---|---|---|
| systemd | Debian, Ubuntu, Fedora, Arch, Raspberry Pi OS | `/etc/systemd/system/loger.service` | `systemctl daemon-reload` e `systemctl enable --now loger` |
| OpenRC | Alpine, Gentoo, Artix | `/etc/init.d/loger` | `rc-update add loger default` e `rc-service loger start` |
| runit | Void, Artix | `/etc/sv/loger/run` (no Artix, `/etc/runit/sv/loger/run`) | link em `/var/service/loger` (ou `/etc/service`, ou `/run/runit/service` no Artix) |

Nos três, o Loger é reiniciado sozinho se cair, e `reload` manda `HUP` para
ele reler a configuração.

Rodar o instalador de novo **não sobrescreve** um `/usr/local/bin/loger` ou um
arquivo de serviço que já existam; ele só garante que o serviço está habilitado.
Para atualizar o Loger, apague o binário (`sudo rm /usr/local/bin/loger`) e
rode o instalador de novo.

Se a instalação do Loger falhar (ou se o init não for nenhum dos três), o
instalador só avisa: os outros serviços continuam instalados normalmente. Com
outro init (s6, SysVinit…), deixe `loger daemon` rodando como root pelo init
da distro.

## Consultar os eventos

```bash
sudo loger show -n 50          # últimos 50 eventos
sudo loger show -f             # em tempo real
sudo loger show -k session     # quem logou, de onde e por quanto tempo
sudo loger show -l alert -s 7d # só alertas da última semana
sudo loger report              # resumo das últimas 24h
sudo loger verify              # confere se alguém mexeu nos logs
systemctl status loger         # estado do serviço (systemd)
rc-service loger status        # OpenRC
sv status loger                # runit
```

## Configuração (opcional)

Funciona sem configuração. Para gerar um arquivo e ajustar:

```bash
sudo mkdir -p /etc/loger
loger config | sudo tee /etc/loger/loger.toml
sudo systemctl reload loger    # OpenRC: sudo rc-service loger reload · runit: sudo sv hup loger
```

Mudanças em `[log]`, `[remote]` ou em quais coletores estão ligados só valem
depois de reiniciar o serviço (`sudo systemctl restart loger`,
`sudo rc-service loger restart` ou `sudo sv restart loger`).

## Desinstalar e dados

Na opção "Desinstalar" do instalador, se o Loger estiver instalado, ele
pergunta `Também desinstalar o Loger? [s/N]`. Respondendo `s`, ele para o
serviço e remove o arquivo de serviço (de qualquer um dos três inits) e o
binário.

Os logs em `/var/log/loger` **não são apagados**: eles recebem `chattr +a`
(append-only) e servem de evidência. Para removê-los manualmente:

```bash
sudo chattr -R -a /var/log/loger && sudo rm -rf /var/log/loger
```

## Envio remoto

Se alguém virar root no servidor, pode apagar os logs locais. O Loger pode
copiar cada evento, cifrado, para outra máquina (receptor). O receptor não é
instalado por este repositório; veja a seção "Envio remoto" no
[README do projeto](https://github.com/DeathHapyness/Loger).
