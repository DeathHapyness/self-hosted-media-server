const log = document.getElementById("log");
const form = document.getElementById("form");
const campo = document.getElementById("campo");
const botao = form.querySelector("button");
const rotulo = document.getElementById("rotulo");
let terminou = false;

function adicionar(tipo, texto) {
  texto = texto.replace(/^\n+|\n+$/g, "");
  if (!texto) return;
  const linha = document.createElement("div");
  linha.className = tipo;
  linha.textContent = texto;
  log.appendChild(linha);
  window.scrollTo(0, document.body.scrollHeight);
}

function habilitar(sim, texto) {
  campo.disabled = botao.disabled = !sim;
  rotulo.textContent = texto;
  if (sim) campo.focus();
}

const fonte = new EventSource("/eventos");
fonte.onopen = () => { if (campo.disabled) rotulo.textContent = "Aguardando o instalador…"; };
fonte.onerror = () => {
  if (terminou) return;
  habilitar(false, "Conexão perdida. Tentando reconectar…");
};
fonte.onmessage = (e) => {
  const m = JSON.parse(e.data);
  adicionar(m.tipo, m.texto);
  if (m.tipo === "pergunta") {
    habilitar(true, "Digite a resposta e pressione Enter (vazio aceita o padrão).");
  } else if (m.tipo === "resposta") {
    habilitar(false, "Executando…");
  } else if (m.tipo === "fim") {
    terminou = true;
    fonte.close();
    form.hidden = true;
  }
};

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  if (campo.disabled) return;
  const resposta = campo.value;
  campo.value = "";
  habilitar(false, "Enviando…");
  try {
    const r = await fetch("/responder", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ resposta }),
    });
    if (!r.ok && r.status !== 409) throw new Error(r.status);
  } catch {
    campo.value = resposta;
    habilitar(true, "Falha ao enviar. Tente novamente.");
  }
});
