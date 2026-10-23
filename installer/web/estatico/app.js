const $ = (id) => document.getElementById(id);

function el(tag, classe, texto) {
  const e = document.createElement(tag);
  if (classe) e.className = classe;
  if (texto != null) e.textContent = texto;
  return e;
}

const FLUXOS = {
  instalar: {
    nome: "Instalação",
    etapas: ["Verificar", "Configurar", "Serviços", "Preparar", "Subir"],
    regras: [
      [/^INSTALAÇÃO$/, 0],
      [/Onde este servidor|\.env already present/, 1],
      [/Quais serviços instalar/, 2],
      [/Required ports|Creating directories|Validating compose/, 3],
      [/Starting services/, 4],
    ],
    fim: /^INSTALLATION COMPLETE$/,
  },
  desinstalar: {
    nome: "Desinstalação",
    etapas: ["Parar", "Revisar", "Limpar"],
    regras: [
      [/^DESINSTALAÇÃO$/, 0],
      [/Pastas de dados encontradas/, 1],
      [/ removido$|\[dry-run\] removeria/, 2],
    ],
    fim: null,
  },
};

const INICIO_FLUXO = { "INSTALAÇÃO": "instalar", "DESINSTALAÇÃO": "desinstalar" };
const TELAS = { "MEDIA SERVER INSTALLER": "Menu principal", "GESTÃO DO DISCO DE MÍDIA": "Disco de mídia" };

const estado = {
  fluxo: null,
  etapa: -1,
  concluido: false,
  erro: false,
  titulo: "",
  bloco: [],
  terminou: false,
};

const log = $("log");
const titulo = $("titulo");
const sobretitulo = $("fluxo");
const contexto = $("contexto");
const controles = $("controles");
const conexao = $("conexao");

function limpar(texto) {
  return texto.replace(/^\n+|\n+$/g, "");
}

const MARCAS = { ok: "✓", falha: "✗", aviso: "!", perigo: "!" };
let ultimoNoLog = null;

function entraNoLog(m, texto) {
  if (!texto) return false;
  switch (m.tipo) {
    case "ok":
    case "falha":
    case "aviso":
    case "perigo":
    case "secao":
    case "texto":
      return true;
    case "cabecalho":
      return !TELAS[texto];
    case "info":
      return texto.startsWith("[dry-run]") || ultimoNoLog === "falha" || ultimoNoLog === "detalhe";
    default:
      return false;
  }
}

function adicionarLog(m) {
  const texto = limpar(m.texto).trim();
  if (!entraNoLog(m, texto)) {
    if (m.tipo !== "info") ultimoNoLog = null;
    return;
  }
  const classe = m.tipo === "info" && !texto.startsWith("[dry-run]") ? "detalhe" : m.tipo;
  ultimoNoLog = classe;

  const perto = log.scrollHeight - log.scrollTop - log.clientHeight < 40;
  const linha = el("div", classe);
  linha.append(el("time", null, m.hora || ""), el("span", "marca-log", MARCAS[m.tipo] || ""), el("span", null, texto));
  log.appendChild(linha);
  if (perto) log.scrollTop = log.scrollHeight;
}

function iniciarFluxo(nome) {
  Object.assign(estado, { fluxo: nome, etapa: 0, concluido: false, erro: false });
}

function avancarEtapa(texto) {
  const fluxo = FLUXOS[estado.fluxo];
  if (!fluxo) return;
  for (const [regra, indice] of fluxo.regras) {
    if (regra.test(texto)) estado.etapa = Math.max(estado.etapa, indice);
  }
  if (fluxo.fim && fluxo.fim.test(texto)) estado.concluido = true;
  if (/Instalação interrompida/.test(texto)) estado.erro = true;
}

function desenharEtapas() {
  const nav = $("etapas");
  const fluxo = FLUXOS[estado.fluxo];
  nav.hidden = !fluxo;
  if (!fluxo) return;

  const lista = nav.querySelector("ol");
  lista.replaceChildren();
  fluxo.etapas.forEach((nome, i) => {
    const item = el("li");
    const feita = estado.concluido || i < estado.etapa;
    if (feita) item.classList.add("feita");
    else if (i === estado.etapa) item.classList.add(estado.erro ? "erro" : "atual");
    const bola = el("span", "bola", feita ? "✓" : estado.erro && i === estado.etapa ? "✗" : String(i + 1));
    item.append(bola, el("span", null, nome));
    lista.appendChild(item);
  });
}

function desenharCabecalho() {
  const fluxo = FLUXOS[estado.fluxo];
  sobretitulo.textContent = fluxo
    ? `${fluxo.nome} · etapa ${Math.min(estado.etapa + 1, fluxo.etapas.length)} de ${fluxo.etapas.length}`
    : "";
  if (fluxo && estado.concluido) sobretitulo.textContent = `${fluxo.nome} · concluída`;
  if (fluxo && estado.erro) sobretitulo.textContent = `${fluxo.nome} · interrompida`;
  titulo.textContent = estado.titulo || "Instalador";
}

function desenharContexto(linhas) {
  contexto.replaceChildren(...linhas.map((m) => el("p", m.tipo, limpar(m.texto))));
}

function mostrarExecutando(texto) {
  const linha = el("div", "executando");
  linha.append(el("span", "girando"), el("span", null, texto || "Executando…"));
  controles.replaceChildren(linha);
}

async function enviar(resposta) {
  for (const c of controles.querySelectorAll("button, input")) c.disabled = true;
  try {
    const r = await fetch("/responder", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ resposta }),
    });
    if (!r.ok && r.status !== 409) throw new Error(r.status);
  } catch {
    for (const c of controles.querySelectorAll("button, input")) c.disabled = false;
    controles.appendChild(el("p", "dica", "Falha ao enviar. Tente novamente."));
  }
}

function botao(texto, classe, aoClicar) {
  const b = el("button", classe ? `botao ${classe}` : "botao", texto);
  b.type = "button";
  b.addEventListener("click", aoClicar);
  return b;
}

function desenharServicos(opcoes) {
  const grade = el("div", "servicos");
  const caixas = [];
  for (const op of opcoes) {
    const linha = el("div", "servico");
    const rotulo = el("label");
    const caixa = el("input");
    caixa.type = "checkbox";
    caixa.checked = true;
    caixa.value = op.valor;
    caixas.push(caixa);
    rotulo.append(caixa, el("span", null, op.rotulo));
    const editar = el("button", "link", "editar");
    editar.type = "button";
    editar.addEventListener("click", () => abrirEditor(op.rotulo));
    linha.append(rotulo, editar);
    grade.appendChild(linha);
  }

  const dica = el("p", "dica", "");
  const marcar = (v) => caixas.forEach((c) => (c.checked = v));
  const barra = el("div", "barra");
  barra.append(
    botao("Todos", "secundario", () => marcar(true)),
    botao("Nenhum", "secundario", () => marcar(false)),
    el("span", "espaco"),
    botao("Continuar →", null, () => {
      const marcados = caixas.filter((c) => c.checked).map((c) => c.value);
      if (!marcados.length) {
        dica.textContent = "Marque pelo menos um serviço.";
        return;
      }
      enviar(marcados.length === caixas.length ? "" : marcados.join(","));
    }),
  );
  controles.replaceChildren(grade, barra, dica);
}

function desenharOpcoes(opcoes, padrao, cancelavel) {
  const grade = el("div", "opcoes");
  for (const op of opcoes) {
    const b = el("button", op.valor === padrao ? "opcao padrao" : "opcao");
    b.type = "button";
    const texto = el("span", null, op.rotulo);
    if (op.valor === padrao) texto.prepend(el("span", "etiqueta", "padrão"));
    b.append(el("span", "num", op.valor), texto);
    b.addEventListener("click", () => enviar(op.valor));
    grade.appendChild(b);
  }
  controles.replaceChildren(grade);
  if (cancelavel) controles.appendChild(el("div", "barra")).appendChild(botao("Cancelar", "secundario", () => enviar("")));
}

function desenharSimNao(pergunta, padrao) {
  const barra = el("div", "barra");
  const sim = botao("Sim", padrao === "s" ? null : "secundario", () => enviar("s"));
  const nao = botao("Não", padrao === "n" ? null : "secundario", () => enviar("n"));
  barra.append(sim, nao);
  controles.replaceChildren(el("p", null, pergunta), barra);
}

function desenharTexto(pergunta, padrao, palavra) {
  const campo = el("input", palavra ? "campo perigo" : "campo");
  campo.autocomplete = "off";
  if (padrao) campo.placeholder = padrao;
  const enviarCampo = () => enviar(campo.value);
  campo.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      enviarCampo();
    }
  });

  const barra = el("div", "barra");
  barra.append(campo, botao(palavra ? "Confirmar" : "Continuar →", palavra ? "perigo" : null, enviarCampo));
  campo.style.flex = "1";

  let dica = "";
  if (palavra) dica = `Digite exatamente "${palavra}" para confirmar. Qualquer outra coisa cancela.`;
  else if (padrao) dica = `Deixe vazio para usar ${padrao}.`;

  const filhos = [];
  if (pergunta) filhos.push(el("p", null, pergunta));
  filhos.push(barra);
  if (dica) filhos.push(el("p", "dica", dica));
  controles.replaceChildren(...filhos);
  campo.focus();
}

function desenharPergunta(prompt) {
  const opcoes = [];
  const linhas = [];
  let palavra = null;

  for (const m of estado.bloco) {
    const texto = limpar(m.texto);
    const op = m.tipo === "info" && texto.match(/^(\d+)\)\s+(.+)$/);
    const confirmar = texto.match(/Digite exatamente "([^"]+)"/);
    if (op) {
      opcoes.push({ valor: op[1], rotulo: op[2] });
    } else if (confirmar) {
      palavra = confirmar[1];
    } else if (["info", "aviso", "falha", "perigo", "texto"].includes(m.tipo) && texto) {
      linhas.push(m);
    }
  }

  const bruto = limpar(prompt).trim();
  const entreColchetes = bruto.match(/\[([^\]]*)\]:?\s*$/);
  const pergunta = bruto.replace(/\s*\[[^\]]*\]:?\s*$/, "").replace(/:\s*$/, "").replace(/^>$/, "");

  if (!estado.titulo) estado.titulo = pergunta;
  desenharCabecalho();
  desenharContexto(linhas);

  let padrao = null;
  for (const op of opcoes) {
    if (/\(padrão\)/.test(op.rotulo)) {
      op.rotulo = op.rotulo.replace(/\s*\(padrão\)/, "");
      padrao = op.valor;
    }
  }
  if (entreColchetes && opcoes.some((o) => o.valor === entreColchetes[1])) padrao = entreColchetes[1];

  if (/separados por vírgula/i.test(bruto) && opcoes.length) {
    desenharServicos(opcoes.filter((o) => o.valor !== "0"));
  } else if (opcoes.length) {
    desenharOpcoes(opcoes, padrao, /vazio para cancelar/i.test(bruto));
  } else if (/\[s\/n\]/i.test(bruto)) {
    desenharSimNao(pergunta, entreColchetes[1].includes("S") ? "s" : "n");
  } else if (/Pressione Enter/i.test(bruto)) {
    controles.replaceChildren(botao("Voltar ao menu", null, () => enviar("")));
  } else {
    const padraoTexto = (entreColchetes && entreColchetes[1]) || (bruto.match(/\(padrão ([^)]+)\)/) || [])[1];
    desenharTexto(pergunta, padraoTexto, palavra);
  }
}

function processar(m) {
  adicionarLog(m);
  const texto = limpar(m.texto).trim();

  switch (m.tipo) {
    case "cabecalho":
      estado.bloco = [];
      if (INICIO_FLUXO[texto]) {
        iniciarFluxo(INICIO_FLUXO[texto]);
        estado.titulo = FLUXOS[estado.fluxo].etapas[0];
      } else if (TELAS[texto]) {
        estado.fluxo = null;
        estado.titulo = TELAS[texto];
      } else {
        estado.titulo = texto.charAt(0) + texto.slice(1).toLowerCase();
      }
      avancarEtapa(texto);
      break;
    case "secao":
      estado.bloco = [];
      estado.titulo = texto.replace(/:$/, "");
      avancarEtapa(texto);
      break;
    case "pergunta":
      if (estado.fluxo && !estado.erro && /Pressione Enter/i.test(texto)) estado.concluido = true;
      desenharEtapas();
      desenharPergunta(m.texto);
      return;
    case "resposta":
      estado.bloco = [];
      contexto.replaceChildren();
      mostrarExecutando();
      return;
    case "fim":
      estado.terminou = true;
      fonte.close();
      sobretitulo.textContent = "Sessão encerrada";
      titulo.textContent = "Tudo pronto";
      contexto.replaceChildren(el("p", null, texto));
      controles.replaceChildren();
      conexao.textContent = "Servidor encerrado";
      conexao.className = "pilula";
      $("abrir-editor").hidden = true;
      return;
    default:
      estado.bloco.push(m);
      avancarEtapa(texto);
  }

  desenharEtapas();
  desenharCabecalho();
  if (!controles.querySelector("input, .opcao, .botao") || controles.querySelector(".executando")) {
    mostrarExecutando(m.tipo === "cabecalho" || m.tipo === "secao" ? null : texto);
  }
}

const fonte = new EventSource("/eventos");
fonte.onopen = () => {
  conexao.textContent = "Conectado";
  conexao.className = "pilula online";
};
fonte.onerror = () => {
  if (estado.terminou) return;
  conexao.textContent = "Reconectando…";
  conexao.className = "pilula offline";
};
fonte.onmessage = (e) => processar(JSON.parse(e.data));

const editor = {
  dialogo: $("editor"),
  servico: $("editor-servico"),
  texto: $("editor-texto"),
  destaque: $("editor-destaque"),
  status: $("editor-status"),
  original: "",
  atual: null,
  carregado: false,
};

function statusEditor(texto, tipo) {
  editor.status.textContent = texto;
  editor.status.className = tipo ? `editor-status ${tipo}` : "editor-status";
}

function escapar(texto) {
  return texto.replace(/[&<>]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;" })[c]);
}

function marcar(classe, texto) {
  return texto ? `<span class="y-${classe}">${escapar(texto)}</span>` : "";
}

const TOKENS_VALOR = /(\s+#.*$)|("(?:[^"\\]|\\.)*"?|'[^']*'?)|(\$\{[^}]*\}?)|([&*][\w-]+)|((?:true|false|yes|no|null|on|off|~)(?=\s|$))|(-?\d+(?:\.\d+)?(?=\s|$))|([|>][-+]?(?=\s|$))|([^\s"'$&*#]+|\s+|.)/gi;

function destacarValor(valor) {
  let html = "";
  for (const m of valor.matchAll(TOKENS_VALOR)) {
    const [, comentario, texto, variavel, ancora, literal, numero, bloco] = m;
    if (comentario) html += marcar("comentario", comentario);
    else if (texto) html += marcar("texto", texto);
    else if (variavel) html += marcar("variavel", variavel);
    else if (ancora) html += marcar("variavel", ancora);
    else if (literal) html += marcar("literal", literal);
    else if (numero) html += marcar("numero", numero);
    else if (bloco) html += marcar("pontuacao", bloco);
    else if (m[0].trim()) html += marcar("texto", m[0]);
    else html += escapar(m[0]);
  }
  return html;
}

function destacarLinha(linha) {
  const [, recuo, tracos, resto] = linha.match(/^(\s*)((?:-(?:\s+|$))*)(.*)$/);
  let html = escapar(recuo) + marcar("pontuacao", tracos);
  if (resto.startsWith("#")) return html + marcar("comentario", resto);

  const chave = resto.match(/^("[^"]*"|'[^']*'|[^\s"'#][^#]*?)(:)(?=\s|$)(.*)$/);
  if (chave) return html + marcar("chave", chave[1]) + marcar("pontuacao", chave[2]) + destacarValor(chave[3]);
  return html + destacarValor(resto);
}

function destacar() {
  editor.destaque.innerHTML = editor.texto.value.split("\n").map(destacarLinha).join("\n") + "\n ";
  sincronizarRolagem();
}

function sincronizarRolagem() {
  editor.destaque.scrollTop = editor.texto.scrollTop;
  editor.destaque.scrollLeft = editor.texto.scrollLeft;
}

editor.texto.addEventListener("input", destacar);
editor.texto.addEventListener("scroll", sincronizarRolagem);

const alterado = () => editor.atual !== null && editor.texto.value !== editor.original;

async function carregarServicos() {
  if (editor.carregado) return;
  const r = await fetch("/api/compose");
  const nomes = await r.json();
  editor.servico.replaceChildren(...nomes.map((n) => {
    const op = el("option", null, n);
    op.value = n;
    return op;
  }));
  editor.carregado = true;
}

async function carregarCompose(nome) {
  statusEditor("Carregando…");
  const r = await fetch(`/api/compose/${encodeURIComponent(nome)}`);
  if (!r.ok) {
    statusEditor(await r.text(), "falha");
    return;
  }
  editor.original = await r.text();
  editor.texto.value = editor.original;
  destacar();
  editor.atual = nome;
  editor.servico.value = nome;
  statusEditor("");
}

async function abrirEditor(nome) {
  await carregarServicos();
  if (!editor.dialogo.open) editor.dialogo.showModal();
  const alvo = nome || editor.atual || editor.servico.value;
  if (alvo !== editor.atual) {
    if (alterado() && !confirm("Descartar as alterações não salvas?")) return;
    await carregarCompose(alvo);
  }
  editor.texto.focus();
}

async function enviarCompose(metodo, url) {
  statusEditor(metodo === "PUT" ? "Validando e salvando…" : "Validando…");
  const r = await fetch(url, { method: metodo, body: editor.texto.value });
  if (!r.ok) {
    statusEditor(await r.text(), "falha");
    return null;
  }
  const resultado = await r.json();
  statusEditor(resultado.mensagem, resultado.ok ? "ok" : "falha");
  return resultado;
}

editor.servico.addEventListener("change", () => {
  const nome = editor.servico.value;
  if (alterado() && !confirm("Descartar as alterações não salvas?")) {
    editor.servico.value = editor.atual;
    return;
  }
  carregarCompose(nome);
});

$("editor-validar").addEventListener("click", () =>
  enviarCompose("POST", `/api/compose/${encodeURIComponent(editor.atual)}/validar`));

async function salvarCompose() {
  const resultado = await enviarCompose("PUT", `/api/compose/${encodeURIComponent(editor.atual)}`);
  if (resultado && resultado.ok) editor.original = editor.texto.value;
}
$("editor-salvar").addEventListener("click", salvarCompose);

$("editor-desfazer").addEventListener("click", () => {
  editor.texto.value = editor.original;
  destacar();
  statusEditor("Alterações desfeitas.");
});

editor.texto.addEventListener("keydown", (e) => {
  if (e.key === "Tab") {
    e.preventDefault();
    editor.texto.setRangeText("  ", editor.texto.selectionStart, editor.texto.selectionEnd, "end");
    destacar();
  } else if ((e.ctrlKey || e.metaKey) && e.key === "s") {
    e.preventDefault();
    salvarCompose();
  }
});

editor.dialogo.addEventListener("cancel", (e) => {
  if (alterado() && !confirm("Fechar sem salvar?")) e.preventDefault();
});
editor.dialogo.querySelector(".fechar").addEventListener("click", (e) => {
  if (alterado() && !confirm("Fechar sem salvar?")) e.preventDefault();
});

$("abrir-editor").addEventListener("click", () => abrirEditor());
