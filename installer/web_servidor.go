package installer

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const validadeToken = 15 * time.Minute

func ipLocal() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func ModoWeb(porta int) error {
	w := novaWeb()
	ouvinte, err := net.Listen("tcp", fmt.Sprintf(":%d", porta))
	if err != nil {
		return fmt.Errorf(Tr("não foi possível abrir a porta %d para a interface web: %w", "could not open port %d for the web interface: %w"), porta, err)
	}
	servidor := &http.Server{Handler: w.rotas(), ReadHeaderTimeout: 10 * time.Second}
	go servidor.Serve(ouvinte)

	porta = ouvinte.Addr().(*net.TCPAddr).Port
	slog.Info("Web interface started", "port", porta)
	Registrar("web_mode", "port", porta)

	PrintOk(Tr("Interface web iniciada.", "Web interface started."))
	PrintInfo(Tr("Abra este link no navegador (vale para um único acesso):", "Open this link in your browser (single use):"))
	PrintInfo(ColorText(fmt.Sprintf("http://%s:%d/?token=%s", ipLocal(), porta, w.token), BOLD))
	PrintInfo(fmt.Sprintf(Tr("O link expira em %d minutos se não for usado. Ctrl+C cancela.", "The link expires in %d minutes if unused. Ctrl+C cancels."),
		int(validadeToken.Minutes())))

	time.AfterFunc(validadeToken, func() {
		w.mu.Lock()
		naoUsado := w.token != ""
		w.token = ""
		w.mu.Unlock()
		if naoUsado {
			Terminal{}.Aviso(Tr("O link da interface web expirou sem ser usado.", "The web interface link expired without being used."))
			slog.Warn("Web token expired unused")
			w.fechar()
		}
	})

	UI = w
	aoTrocarIdioma = func() { w.emitir("idioma", IDIOMA) }
	MainMenu()
	aoTrocarIdioma = nil
	UI = Terminal{}

	w.emitir("fim", Tr("Configuração concluída. O servidor web foi encerrado; pode fechar esta aba.", "Setup finished. The web server was shut down; you can close this tab."))
	w.fechar()
	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()
	if err := servidor.Shutdown(ctx); err != nil {
		slog.Warn("Web server did not shut down cleanly", "error", err)
	}
	slog.Info("Web interface closed")
	PrintOk(Tr("Interface web encerrada.", "Web interface closed."))
	return nil
}
