# Modo web do instalador

Além do terminal, o instalador pode ser usado pelo navegador. É o mesmo
instalador, com os mesmos menus e as mesmas perguntas: a diferença é só onde
você responde. Em vez de digitar no terminal, você clica em botões e preenche
campos em uma página web servida pelo próprio instalador.

Isso é útil principalmente quando o servidor não tem monitor (headless) e você
quer configurá-lo a partir do notebook ou do celular, na mesma rede.

## Aviso importante: a conexão NÃO usa HTTPS

<span style="color: red; font-weight: bold;">A interface web funciona apenas em HTTP, sem criptografia.</span>

Na prática, isso significa:

- Tudo o que trafega entre o navegador e o servidor (o link de acesso, o
  cookie da sessão, as respostas que você digita e o conteúdo dos
  `docker-compose.yml`) passa pela rede **em texto puro**.
- Qualquer pessoa ou dispositivo na mesma rede que consiga observar o tráfego
  (por exemplo, em um Wi-Fi público ou compartilhado) poderia ler esses dados
  e, em tese, usar a sessão aberta.
- O servidor web escuta em **todas as interfaces de rede** da máquina, não só
  em `localhost`. Ele fica acessível para qualquer dispositivo que alcance o IP
  e a porta do servidor.

Por isso, siga estas recomendações:

- Use o modo web **somente na sua rede doméstica**, em que você confia. Nunca
  em redes públicas.
- **Não exponha a porta** do instalador (`8787` por padrão) para a internet
  (sem redirecionamento de porta no roteador).
- Se precisar configurar de fora de casa, use uma VPN como o
  [Tailscale](tailscale.md) e acesse pelo IP da VPN, que já é criptografado.
- Se estiver em uma rede em que não confia, prefira o modo terminal (por SSH),
  que é criptografado.

O instalador reduz esse risco com um link de uso único, que expira e é
encerrado automaticamente (veja abaixo), mas isso **não substitui** HTTPS.

## Como usar

1. Rode o instalador normalmente, com `sudo`:

   ```bash
   sudo ./instalador-linux-amd64
   ```

2. Na tela **MODO DE INSTALAÇÃO**, escolha a opção `2) Navegador (Web)`.
   Apertar Enter sem digitar nada escolhe o terminal, que é o padrão.

3. O terminal mostra um link parecido com este:

   ```
   http://192.168.0.10:8787/?token=ABCDEFGHIJKLMNOPQRSTUVWXYZ
   ```

   O IP mostrado é o IP da máquina na rede local, detectado automaticamente.

4. Abra o link no navegador de qualquer dispositivo na mesma rede.

5. A partir daqui, use a página. O terminal fica parado mostrando o link,
   esperando a interface web terminar.

6. Quando você escolher **Sair** no menu da página, o instalador encerra o
   servidor web, a página mostra "Tudo pronto" e você pode fechar a aba.

### Escolhendo outra porta

A porta padrão é `8787`. Para usar outra, passe `--web-porta`:

```bash
sudo ./instalador-linux-amd64 --web-porta 9000
```

Com `--web-porta 0`, o sistema escolhe uma porta livre qualquer, e o link
mostrado no terminal já vem com ela.

Se a porta estiver ocupada, o instalador avisa e encerra com erro, sem abrir a
interface.

## Como funciona o acesso (link de uso único)

O link contém um **token**: um código aleatório gerado a cada execução.

- **Vale uma única vez.** No primeiro acesso, o token é consumido e o
  navegador recebe um cookie de sessão. Depois disso, o link deixa de
  funcionar: abrir de novo (ou em outro dispositivo) mostra
  "Link inválido ou já utilizado."
- **Expira em 15 minutos se não for usado.** Se ninguém abrir o link nesse
  tempo, o instalador avisa no terminal que o link expirou e encerra. Para
  tentar de novo, rode o instalador outra vez e um link novo será gerado.
- **Só o navegador que usou o link tem acesso.** Todas as páginas e ações
  exigem o cookie de sessão. Sem ele, a resposta é "Acesso negado".
- **Recarregar a página funciona.** Como o cookie fica no navegador, você pode
  recarregar a aba sem perder a sessão. O histórico do log é reenviado.
- **Tudo acaba ao sair.** Quando você sai pelo menu, o servidor web é
  desligado. Não fica nenhum serviço rodando em segundo plano.

Para cancelar a qualquer momento, aperte `Ctrl+C` no terminal onde o
instalador está rodando.

Cada acesso feito pelo link fica registrado no log
(`/var/log/media-server-installer.log`) e no histórico em JSON, com o endereço
de origem.

## O que aparece na página

A página tem três partes:

- **Cartão central**: mostra a pergunta atual do instalador e os controles para
  respondê-la. Por exemplo, botões para cada opção de menu, botões Sim/Não,
  uma lista de serviços para marcar ou um campo de texto. Quando o instalador
  está trabalhando, o cartão mostra o que está sendo executado.
- **Etapas**: durante a instalação e a desinstalação, uma barra mostra em que
  etapa você está (por exemplo, Verificar, Configurar, Serviços, Preparar e
  Subir).
- **Log**: a coluna lateral mostra, em tempo real, tudo o que o instalador
  imprimiria no terminal: sucessos, avisos, falhas e as respostas que você
  deu.

No canto superior há um indicador de conexão:

- **Conectado**: a página está recebendo as atualizações do instalador.
- **Reconectando…**: a conexão caiu (por exemplo, o Wi-Fi oscilou). A página
  tenta reconectar sozinha e continua de onde parou.
- **Servidor encerrado**: a sessão terminou normalmente.

Confirmações perigosas, como apagar pastas de dados ou formatar um disco,
continuam pedindo que você digite a palavra exata (por exemplo, `CONFIRMO`),
igual ao terminal.

## Editor de docker-compose

No topo da página há o botão **editar compose**. Ele abre um editor para os
arquivos `docker-compose.yml` dos serviços do repositório.

Como usar:

1. Escolha o serviço na lista no topo do editor.
2. Edite o texto. A tecla `Tab` insere dois espaços, como o YAML espera.
3. Clique em **Validar** para conferir o arquivo sem salvar.
4. Clique em **Salvar** (ou `Ctrl+S`) para gravar.
5. **Desfazer** volta ao conteúdo que estava salvo.

O que acontece ao validar e salvar:

- **A validação usa o próprio Compose.** O conteúdo é gravado em um arquivo
  temporário (`.docker-compose.validar.yml`, na pasta do serviço) e conferido
  com `docker compose config -q`. Em seguida, o temporário é apagado. Se
  houver erro, a mensagem do Compose aparece no editor.
- **Nada inválido é salvo.** O botão Salvar sempre valida antes. Se a
  validação falhar, o arquivo original não é tocado.
- **Sempre fica um backup.** Antes de gravar, o arquivo atual é copiado para
  `docker-compose.yml.bak`, na mesma pasta. Para voltar atrás, basta renomear o
  `.bak` de volta. Atenção: cada salvamento sobrescreve o `.bak` anterior, então
  só a última versão antes da edição fica guardada.
- **As permissões do arquivo são mantidas.**
- **Com `--dry-run`, nada é gravado.** O editor valida e avisa que nada foi
  salvo.
- Toda edição salva fica registrada no log e no histórico em JSON.

Observações:

- Salvar o arquivo **não reinicia** o serviço. Para aplicar a mudança, rode a
  opção "Instalar / atualizar serviços" de novo (ela executa
  `docker compose up -d`, que recria só os containers que mudaram).
- Se você fechar o editor ou trocar de serviço com alterações não salvas, a
  página pede confirmação antes de descartá-las.
- Os arquivos alterados fazem parte do repositório. Se você atualizar o
  repositório com `git pull`, suas mudanças podem gerar conflito. Use
  `git status` para ver o que foi alterado.

## Diferenças em relação ao terminal

| | Terminal | Web |
|---|---|---|
| Menus e perguntas | Iguais | Iguais (mostrados como botões e campos) |
| Criptografia | Sim, se usado por SSH | Não (HTTP puro) |
| Editor de docker-compose | Não | Sim |
| Precisa de rede | Não | Sim, o navegador precisa alcançar o servidor |
| Cancelar | `Ctrl+C` | `Ctrl+C` no terminal do instalador |

## Limitações conhecidas

- **Sem HTTPS**, como explicado no início desta página.
- **Fechar a aba não encerra o instalador.** Se você fechar o navegador no meio
  de uma pergunta, o instalador continua esperando a resposta. Como o link já
  foi usado, não dá para abrir uma nova sessão em outro dispositivo; volte ao
  mesmo navegador (o cookie continua válido) ou encerre com `Ctrl+C` no
  terminal e rode de novo.
- **Uma sessão por execução.** Só um navegador por vez controla o instalador.

## Solução de problemas

**"Link inválido ou já utilizado."**
O token já foi consumido (talvez por uma pré-visualização do link em um
aplicativo de mensagens) ou expirou. Encerre o instalador com `Ctrl+C` e rode de
novo para gerar outro link.

**"Acesso negado. Use o link exibido no terminal."**
Você abriu o endereço sem o token, ou em um navegador diferente do que usou o
link. Use o mesmo navegador ou gere um novo link.

**A página não abre.**
- Confira se o dispositivo está na mesma rede que o servidor.
- Confira se um firewall (por exemplo, `ufw`) está bloqueando a porta. Para
  liberar temporariamente: `sudo ufw allow 8787/tcp`. Lembre-se de remover a
  regra depois: `sudo ufw delete allow 8787/tcp`.
- Se o IP mostrado for `localhost`, o instalador não conseguiu detectar o IP da
  rede. Descubra o IP com `ip -4 addr` e troque no link.

**"não foi possível abrir a porta 8787"**
Outro programa já usa essa porta. Use `--web-porta` com outra porta, ou
`--web-porta 0`.
