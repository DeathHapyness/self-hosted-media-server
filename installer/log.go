package installer

import (
	"log/slog"
	"os"
)

const LOG_FILE = "/var/log/media-server-installer.log"

func SetupLogging(loFile string) error {
	arquivo, err := os.OpenFile(loFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	handler := slog.NewJSONHandler(arquivo, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: renomearCampos,
	})
	slog.SetDefault(slog.New(handler))
	return nil

}

func renomearCampos(grupos []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		return slog.String("hora", a.Value.Time().Format("2006-01-02 15:04:05"))
	case slog.LevelKey:
		a.Key = "nivel"
	case slog.MessageKey:
		a.Key = "mensagem"
	}
	return a
}
