package installer

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	Execucoes []Execucao `json:"execucoes"`
}

var historico = Historico{Execucoes: []Execucao{}}
var atual *Execucao // equivale ao Optional[dict] = None

var HistoryFile = filepath.Join(BASE_DIR, "json", "media-server-history.json")

func historyPath() string {
	exe, err := os.Executable()
	if err != nil {
		return filepath.Join("json", "media-server-history.json")
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	return filepath.Join(filepath.Dir(exe), "json", "media-server-history.json")
}

func registrar(acao string, detalhes ...any) {
	slog.Info("historico", append([]any{"acao", acao}, detalhes...)...)
}

func _agora() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func carregar() Historico {
	vazio := Historico{Execucoes: []Execucao{}}
	dados, err := os.ReadFile(HistoryFile)
	if err != nil {
		return Historico{Execucoes: []Execucao{}}
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

func salvar() string {
	tmp := HistoryFile.TrimSuffix(".json.tmp")
}
