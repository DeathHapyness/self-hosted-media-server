package installer

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const MaxExecucoes int = 10000

type Evento struct {
	Hora     string         `json:"time"`
	Acao     string         `json:"action"`
	Detalhes map[string]any `json:"details,omitempty"`
}

type Execucao struct {
	Inicio  string   `json:"start"`
	Fim     *string  `json:"end"`
	Status  string   `json:"status"`
	DryRun  bool     `json:"dry_run"`
	Eventos []Evento `json:"events"`
}

type Historico struct {
	Execucoes []*Execucao `json:"runs"`
}

type historicoLegado struct {
	Execucoes []struct {
		Inicio  string  `json:"inicio"`
		Fim     *string `json:"fim"`
		Status  string  `json:"status"`
		DryRun  bool    `json:"dry_run"`
		Eventos []struct {
			Hora     string         `json:"hora"`
			Acao     string         `json:"acao"`
			Detalhes map[string]any `json:"detalhes,omitempty"`
		} `json:"eventos"`
	} `json:"execucoes"`
}

func (l historicoLegado) converter() Historico {
	h := Historico{Execucoes: []*Execucao{}}
	for _, e := range l.Execucoes {
		ex := &Execucao{Inicio: e.Inicio, Fim: e.Fim, Status: e.Status, DryRun: e.DryRun, Eventos: []Evento{}}
		for _, ev := range e.Eventos {
			ex.Eventos = append(ex.Eventos, Evento{Hora: ev.Hora, Acao: ev.Acao, Detalhes: ev.Detalhes})
		}
		h.Execucoes = append(h.Execucoes, ex)
	}
	return h
}

var historico = Historico{Execucoes: []*Execucao{}}
var atual *Execucao // equivale ao Optional[dict] = None

var HistoryFile = filepath.Join(repoRoot(), "installer", "json", "media-server-history.json")

var historicoMu sync.Mutex

func _agora() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func carregar() Historico {
	vazio := Historico{Execucoes: []*Execucao{}}
	dados, err := os.ReadFile(HistoryFile)
	if err != nil {
		return vazio
	}
	var h Historico
	if err := json.Unmarshal(dados, &h); err == nil && h.Execucoes != nil {
		return h
	}
	var legado historicoLegado
	if err := json.Unmarshal(dados, &legado); err == nil && legado.Execucoes != nil {
		slog.Info("Legacy history format found; migrating", "file", HistoryFile)
		return legado.converter()
	}
	backup := strings.TrimSuffix(HistoryFile, ".json") + ".json.corrupted"
	if err := os.Rename(HistoryFile, backup); err != nil {
		slog.Error("History file unreadable and could not be moved", "error", err)
	} else {
		slog.Warn("History file unreadable; moved aside and started fresh", "backup", backup)
	}
	return vazio
}

func salvar() {
	tmp := HistoryFile + ".tmp"

	if err := os.MkdirAll(filepath.Dir(HistoryFile), 0o755); err != nil {
		slog.Error("Could not write the history", "file", HistoryFile, "error", err)
		return
	}

	dados, err := json.MarshalIndent(historico, "", "  ")
	if err != nil {
		slog.Error("Could not serialize the history", "error", err)
		return
	}

	if err := os.WriteFile(tmp, dados, 0o644); err != nil {
		slog.Error("Could not write the history", "file", HistoryFile, "error", err)
		return
	}

	if err := os.Rename(tmp, HistoryFile); err != nil {
		slog.Error("Could not write the history", "file", HistoryFile, "error", err)
		return
	}
}

func IniciarExecucao() {
	historicoMu.Lock()
	defer historicoMu.Unlock()
	historico = carregar()
	atual = &Execucao{
		Inicio:  _agora(),
		Status:  "running",
		DryRun:  DRY_RUN,
		Eventos: []Evento{},
	}
	historico.Execucoes = append(historico.Execucoes, atual)
	if len(historico.Execucoes) > MaxExecucoes {
		historico.Execucoes = historico.Execucoes[len(historico.Execucoes)-MaxExecucoes:]
	}
	salvar()
	slog.Info("History", "file", HistoryFile)
}

func Registrar(acao string, detalhes ...any) {
	historicoMu.Lock()
	defer historicoMu.Unlock()
	if atual == nil {
		return
	}

	var mapa map[string]any
	if len(detalhes) > 0 {
		mapa = make(map[string]any)
		for i := 0; i+1 < len(detalhes); i += 2 {
			chave, ok := detalhes[i].(string)
			if !ok {
				continue
			}
			mapa[chave] = detalhes[i+1]
		}
	}

	atual.Eventos = append(atual.Eventos, Evento{
		Hora:     _agora(),
		Acao:     acao,
		Detalhes: mapa,
	})
	salvar()
}

func FinalizarExecucao(status string) {
	historicoMu.Lock()
	defer historicoMu.Unlock()
	if atual == nil {
		return
	}
	fim := _agora()
	atual.Fim = &fim
	atual.Status = status
	salvar()
}
