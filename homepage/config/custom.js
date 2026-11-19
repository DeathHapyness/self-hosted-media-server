(() => {
  const GLANCES = `http://${location.hostname}:61208/api/4`;
  const reduzido = matchMedia("(prefers-reduced-motion: reduce)").matches;
  const css = getComputedStyle(document.documentElement);
  const cor = (v) => css.getPropertyValue(v).trim();
  const espera = (ms) => new Promise((r) => setTimeout(r, ms));

  const estadoDo = (card) => {
    const dot = card.querySelector(".site-monitor-status [class*='bg-'], .docker-status [class*='bg-']");
    const c = dot ? dot.className : "";
    if (/emerald|green/.test(c)) return "ok";
    if (/rose|red/.test(c)) return "erro";
    return "aviso";
  };

  const contar = () => {
    const cards = [...document.querySelectorAll(".service-card")].filter((c) => c.querySelector(".site-monitor-status, .docker-status"));
    const e = cards.map(estadoDo);
    return { total: cards.length, ok: e.filter((x) => x === "ok").length, erro: e.filter((x) => x === "erro").length };
  };

  const saudacao = () => {
    const el = document.querySelector(".information-widget-greeting span");
    if (!el || el.dataset.ok) return;
    el.dataset.ok = "1";
    const h = new Date().getHours();
    const s = h < 5 ? "Boa madrugada" : h < 12 ? "Bom dia" : h < 18 ? "Boa tarde" : "Boa noite";
    el.innerHTML = `${s}<em>.</em>`;
  };

  const numerar = () => {
    document.querySelectorAll(".services-list, .bookmark-list").forEach((l) =>
      [...l.children].forEach((c, i) => c.querySelector(".service-card, a, div")?.style.setProperty("--i", i))
    );
  };

  const boot = () => {
    let visto = false;
    try { visto = sessionStorage.getItem("boot") === "1"; sessionStorage.setItem("boot", "1"); } catch {}
    if (visto || reduzido) return;
    const tela = document.createElement("div");
    tela.id = "boot";
    document.body.appendChild(tela);
    const sair = () => { tela.classList.add("saindo"); setTimeout(() => tela.remove(), 600); };
    tela.addEventListener("click", sair);
    addEventListener("keydown", sair, { once: true });
    const linhas = [];
    const escrever = (h) => { linhas.push(h); tela.innerHTML = linhas.join("\n") + '\n<span class="cur"></span>'; };
    (async () => {
      escrever(`media-server · ${new Date().toLocaleString("pt-BR")}`);
      escrever("");
      for (let t = 0; t < 30 && contar().total === 0; t++) await espera(100);
      await espera(900);
      for (const c of document.querySelectorAll(".service-card")) {
        const nome = c.querySelector(".service-name")?.childNodes[0]?.textContent?.trim();
        if (!nome) continue;
        const e = c.querySelector(".site-monitor-status, .docker-status") ? estadoDo(c) : "ok";
        escrever(`${e === "ok" ? '<span class="ok">[ ok ]</span>' : e === "erro" ? '<span class="erro">[fora]</span>' : '<span class="aviso">[ ?? ]</span>'} ${nome}`);
        await espera(28);
      }
      const n = contar();
      escrever("");
      escrever(`${n.ok}/${n.total} serviços respondendo.`);
      await espera(700);
      sair();
    })();
  };

  const barra = document.createElement("div");
  barra.id = "barra";
  barra.innerHTML = `
    <div class="m" title="Serviços monitorados nesta aba"><span class="estado"></span><span data-svc>…</span></div>
    <div class="m" title="Uso de CPU"><span class="r">CPU</span><b data-cpu>–</b><canvas class="spark" data-s="cpu"></canvas></div>
    <div class="m" title="Uso de memória"><span class="r">RAM</span><b data-mem>–</b><canvas class="spark" data-s="mem"></canvas></div>
    <div class="m opc2" title="Temperatura da CPU"><span class="r">TEMP</span><b data-temp>–</b></div>
    <div class="m opc" title="Rede: download / upload"><span class="r">NET</span><span>↓<b data-rx>–</b> ↑<b data-tx>–</b></span><canvas class="spark" data-s="net"></canvas></div>
    <div class="m opc" title="Carga média (1 min) / núcleos"><span class="r">LOAD</span><b data-load>–</b></div>
    <div class="m opc" title="Contêineres rodando"><span class="r">DOCKER</span><b data-ctr>–</b></div>
    <div class="m opc2" title="Tempo ligado"><span class="r">UP</span><b data-up>–</b></div>
    <div class="m fim" title="Última leitura do Glances"><span class="r">ATUALIZADO</span><b data-hora>–</b></div>`;

  const hist = { cpu: [], mem: [], net: [] };
  const q = (s) => barra.querySelector(s);
  let cpu = 0, saude = "ok";

  const tam = (b) => {
    const u = ["B", "K", "M", "G"];
    let i = 0;
    while (b >= 1024 && i < u.length - 1) { b /= 1024; i++; }
    return `${b < 10 && i ? b.toFixed(1) : Math.round(b)}${u[i]}`;
  };

  const nivel = (el, v, a, c) => { el.classList.toggle("alto", v >= a && v < c); el.classList.toggle("critico", v >= c); };

  const spark = (cv, dados, max) => {
    const w = cv.clientWidth, h = cv.clientHeight, dpr = devicePixelRatio || 1;
    cv.width = w * dpr; cv.height = h * dpr;
    const c = cv.getContext("2d");
    c.setTransform(dpr, 0, 0, dpr, 0, 0);
    if (dados.length < 2) return;
    const m = max || Math.max(...dados, 1);
    const pts = dados.map((v, i) => [(i / (dados.length - 1)) * w, h - 1 - (v / m) * (h - 2)]);
    c.beginPath();
    pts.forEach(([x, y], i) => (i ? c.lineTo(x, y) : c.moveTo(x, y)));
    c.strokeStyle = cor("--ac");
    c.lineWidth = 1.2;
    c.stroke();
    c.lineTo(w, h); c.lineTo(0, h); c.closePath();
    c.fillStyle = "rgba(94,234,212,0.10)";
    c.fill();
  };

  const empurrar = (k, v) => { hist[k].push(v); if (hist[k].length > 30) hist[k].shift(); };

  const json = (p) => fetch(`${GLANCES}/${p}`, { cache: "no-store" }).then((r) => r.json());

  const coletar = async () => {
    try {
      const [ql, mem, load, sens, net, up, ctr] = await Promise.all([
        json("quicklook"), json("mem"), json("load"), json("sensors"), json("network"), json("uptime"), json("containers"),
      ]);
      barra.classList.remove("offline");
      cpu = ql.cpu;
      q("[data-cpu]").textContent = `${Math.round(cpu)}%`; nivel(q("[data-cpu]"), cpu, 70, 90);
      q("[data-mem]").textContent = `${Math.round(mem.percent)}%`; nivel(q("[data-mem]"), mem.percent, 75, 90);
      const t = sens.find((s) => /Tctl|Package id 0|CPU|k10temp|coretemp/i.test(s.label) && s.type.includes("temperature")) || sens.find((s) => s.type.includes("temperature"));
      if (t) { q("[data-temp]").textContent = `${Math.round(t.value)}°`; nivel(q("[data-temp]"), t.value, 70, 85); }
      const ifs = net.filter((n) => !/^(lo|veth|br-|docker|tailscale|virbr)/.test(n.interface_name));
      const rx = ifs.reduce((a, n) => a + (n.bytes_recv_rate_per_sec ?? n.rx ?? 0), 0);
      const tx = ifs.reduce((a, n) => a + (n.bytes_sent_rate_per_sec ?? n.tx ?? 0), 0);
      q("[data-rx]").textContent = tam(rx); q("[data-tx]").textContent = tam(tx);
      q("[data-load]").textContent = `${load.min1.toFixed(2)}/${load.cpucore}`; nivel(q("[data-load]"), load.min1 / load.cpucore, 0.8, 1.2);
      q("[data-ctr]").textContent = ctr.filter((c) => /running|healthy|^up/i.test(c.status)).length;
      q("[data-up]").textContent = String(up).replace(/:\d+$/, "").replace(/^(\d+):(\d+)$/, "$1h$2m").replace(/ days?, /, "d ");
      empurrar("cpu", cpu); empurrar("mem", mem.percent); empurrar("net", rx + tx);
      spark(q('[data-s="cpu"]'), hist.cpu, 100);
      spark(q('[data-s="mem"]'), hist.mem, 100);
      spark(q('[data-s="net"]'), hist.net);
      q("[data-hora]").textContent = new Date().toLocaleTimeString("pt-BR");
    } catch {
      barra.classList.add("offline");
    }
  };

  const servicos = () => {
    if (!document.body.contains(barra)) document.body.appendChild(barra);
    const n = contar();
    saude = n.erro > 0 ? "erro" : n.ok < n.total ? "aviso" : "ok";
    q(".estado").className = `estado ${saude === "ok" ? "" : saude}`;
    q("[data-svc]").innerHTML = n.total ? `<b>${n.ok}</b>/${n.total} no ar` : "sem monitor";
  };

  const iniciar = () => {
    if (!document.body) return setTimeout(iniciar, 50);
    boot();
    servicos();
    coletar();
    setInterval(coletar, 2000);
    setInterval(servicos, 3000);
    new MutationObserver(() => requestAnimationFrame(() => { saudacao(); numerar(); })).observe(document.body, { childList: true, subtree: true });
  };
  iniciar();
})();
