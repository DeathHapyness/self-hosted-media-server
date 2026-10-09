package installer

import "log/slog"

func registrar(acao string, detalhes ...any) {
	slog.Info("historico", append([]any{"acao", acao}, detalhes...)...)
}
