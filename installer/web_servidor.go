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
		return fmt.Errorf("não foi possível abrir a porta %d para a interface web: %w", porta, err)
	}
	servidor := &http.Server{Handler: w.rotas(), ReadHeaderTimeout: 10 * time.Second}
	go servidor.Serve(ouvinte)

	porta = ouvinte.Addr().(*net.TCPAddr).Port
	slog.Info("Interface web iniciada", "porta", porta)
	Registrar("modo_web", "porta", porta)

	PrintOk("Interface web iniciada.")
	PrintInfo("Abra este link no navegador (vale para um único acesso):")
	PrintInfo(ColorText(fmt.Sprintf("http://%s:%d/?token=%s", ipLocal(), porta, w.token), BOLD))
	PrintInfo(fmt.Sprintf("O link expira em %d minutos se não for usado. Ctrl+C cancela.",
		int(validadeToken.Minutes())))

	time.AfterFunc(validadeToken, func() {
		w.mu.Lock()
		naoUsado := w.token != ""
		w.token = ""
		w.mu.Unlock()
		if naoUsado {
			Terminal{}.Aviso("O link da interface web expirou sem ser usado.")
			slog.Warn("Token web expirou sem uso")
			w.fechar()
		}
	})

	UI = w
	MainMenu()
	UI = Terminal{}

	w.emitir("fim", "Configuração concluída. O servidor web foi encerrado; pode fechar esta aba.")
	w.fechar()
	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()
	if err := servidor.Shutdown(ctx); err != nil {
		slog.Warn("Servidor web não encerrou limpo", "erro", err)
	}
	slog.Info("Interface web encerrada")
	PrintOk("Interface web encerrada.")
	return nil
}
