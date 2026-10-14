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
	Hora     string         `json:"hora"`
	Acao     string         `json:"acao"`
	Detalhes map[string]any `json:"detalhes,omitempty"`
}

type Execucao struct {
	Inicio  string   `json:"inicio"`
	Fim     *string  `json:"fim"`
	Status  string   `json:"status"`
	DryRun  bool     `json:"dry_run"`
	Eventos []Evento `json:"eventos"`
}

type Historico struct {
	Execucoes []*Execucao `json:"execucoes"`
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
	backup := strings.TrimSuffix(HistoryFile, ".json") + ".json.corrompido"
	if err := os.Rename(HistoryFile, backup); err != nil {
		slog.Error("Historico ilegivel e nao foi possivel move-lo", "error", err)
	} else {
		slog.Warn("Historico ilegivel; Movido e recomecado", "backup", backup)
	}
	return vazio
}

func salvar() {
	tmp := HistoryFile + ".tmp"

	if err := os.MkdirAll(filepath.Dir(HistoryFile), 0o755); err != nil {
		slog.Error("não foi possível gravar o histórico", "arquivo", HistoryFile, "erro", err)
		return
	}

	dados, err := json.MarshalIndent(historico, "", "  ")
	if err != nil {
		slog.Error("não foi possível serializar o histórico", "erro", err)
		return
	}

	if err := os.WriteFile(tmp, dados, 0o644); err != nil {
		slog.Error("não foi possível gravar o histórico", "arquivo", HistoryFile, "erro", err)
		return
	}

	if err := os.Rename(tmp, HistoryFile); err != nil {
		slog.Error("não foi possível gravar o histórico", "arquivo", HistoryFile, "erro", err)
		return
	}
}

func IniciarExecucao() {
	historicoMu.Lock()
	defer historicoMu.Unlock()
	historico = carregar()
	atual = &Execucao{
		Inicio:  _agora(),
		Status:  "em andamento",
		DryRun:  DRY_RUN,
		Eventos: []Evento{},
	}
	historico.Execucoes = append(historico.Execucoes, atual)
	if len(historico.Execucoes) > MaxExecucoes {
		historico.Execucoes = historico.Execucoes[len(historico.Execucoes)-MaxExecucoes:]
	}
	salvar()
	slog.Info("Histórico", "arquivo", HistoryFile)
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
