package installer

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed web
var arquivosWeb embed.FS

const cookieSessao = "instalador_sessao"

var codigosANSI = regexp.MustCompile(`\x1b\[[0-9;]*m`)

var errWebEncerrada = errors.New("web interface closed")

type mensagem struct {
	ID    int    `json:"id"`
	Tipo  string `json:"tipo"`
	Texto string `json:"texto"`
	Hora  string `json:"hora"`
}

type Web struct {
	mu         sync.Mutex
	mensagens  []mensagem
	novidade   chan struct{}
	aguardando bool
	token      string
	sessao     string

	respostas chan string
	encerrada chan struct{}
	encerrar  sync.Once
}

func novaWeb() *Web {
	return &Web{
		novidade:  make(chan struct{}),
		token:     rand.Text(),
		sessao:    rand.Text(),
		respostas: make(chan string, 1),
		encerrada: make(chan struct{}),
	}
}

func (w *Web) emitir(tipo, texto string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.mensagens = append(w.mensagens, mensagem{
		ID:    len(w.mensagens),
		Tipo:  tipo,
		Texto: codigosANSI.ReplaceAllString(texto, ""),
		Hora:  time.Now().Format("15:04:05"),
	})
	close(w.novidade)
	w.novidade = make(chan struct{})
}

func (w *Web) desde(id int) ([]mensagem, chan struct{}) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if id < 0 || id > len(w.mensagens) {
		id = 0
	}
	return append([]mensagem(nil), w.mensagens[id:]...), w.novidade
}

func (w *Web) fechar() {
	w.encerrar.Do(func() { close(w.encerrada) })
}

func (w *Web) Cabecalho(texto string) { w.emitir("cabecalho", texto) }
func (w *Web) Secao(texto string)     { w.emitir("secao", texto) }
func (w *Web) Ok(texto string)        { w.emitir("ok", texto) }
func (w *Web) Falha(texto string)     { w.emitir("falha", texto) }
func (w *Web) Aviso(texto string)     { w.emitir("aviso", texto) }
func (w *Web) Info(texto string)      { w.emitir("info", texto) }
func (w *Web) Perigo(texto string)    { w.emitir("perigo", texto) }
func (w *Web) Texto(texto string)     { w.emitir("texto", texto) }

func (w *Web) Perguntar(prompt string) (string, error) {
	w.mu.Lock()
	w.aguardando = true
	w.mu.Unlock()
	w.emitir("pergunta", prompt)

	select {
	case resposta := <-w.respostas:
		return strings.TrimSpace(resposta), nil
	case <-w.encerrada:
		return "", errWebEncerrada
	}
}

var _ Interface = (*Web)(nil)

func iguais(a, b string) bool {
	return a != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func (w *Web) autorizado(r *http.Request) bool {
	cookie, err := r.Cookie(cookieSessao)
	return err == nil && iguais(cookie.Value, w.sessao)
}

func (w *Web) exigirSessao(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if !w.autorizado(r) {
			http.Error(rw, Tr("Acesso negado.", "Access denied."), http.StatusForbidden)
			return
		}
		h.ServeHTTP(rw, r)
	})
}

func (w *Web) rotas() http.Handler {
	estaticos, _ := fs.Sub(arquivosWeb, "web")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", w.pagina)
	mux.Handle("GET /estatico/", w.exigirSessao(http.FileServerFS(estaticos)))
	mux.HandleFunc("GET /eventos", w.eventos)
	mux.HandleFunc("POST /responder", w.responder)
	mux.Handle("GET /api/compose", w.exigirSessao(http.HandlerFunc(w.listarCompose)))
	mux.Handle("GET /api/compose/{nome}", w.exigirSessao(http.HandlerFunc(w.lerCompose)))
	mux.Handle("POST /api/compose/{nome}/validar", w.exigirSessao(http.HandlerFunc(w.validarCompose)))
	mux.Handle("PUT /api/compose/{nome}", w.exigirSessao(http.HandlerFunc(w.salvarCompose)))

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		h := rw.Header()
		h.Set("Cache-Control", "no-store")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy",
			"default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'")
		mux.ServeHTTP(rw, r)
	})
}

func (w *Web) pagina(rw http.ResponseWriter, r *http.Request) {
	if token := r.URL.Query().Get("token"); token != "" {
		w.mu.Lock()
		valido := iguais(token, w.token)
		if valido {
			w.token = ""
		}
		w.mu.Unlock()

		if !valido {
			slog.Warn("Web access attempt with an invalid or already used token", "origin", r.RemoteAddr)
			http.Error(rw, Tr("Link inválido ou já utilizado.", "Invalid or already used link."), http.StatusForbidden)
			return
		}
		slog.Info("Web token consumed", "origin", r.RemoteAddr)
		Registrar("web_mode", "access", r.RemoteAddr)
		http.SetCookie(rw, &http.Cookie{
			Name:     cookieSessao,
			Value:    w.sessao,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(rw, r, "/", http.StatusSeeOther)
		return
	}

	if !w.autorizado(r) {
		http.Error(rw, Tr("Acesso negado. Use o link exibido no terminal.", "Access denied. Use the link shown in the terminal."), http.StatusForbidden)
		return
	}
	pagina, err := arquivosWeb.ReadFile("web/index.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if IDIOMA == IDIOMA_EN {
		pagina = bytes.Replace(pagina, []byte(`lang="pt-BR"`), []byte(`lang="en"`), 1)
	}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Write(pagina)
}

func escreverEventos(rw http.ResponseWriter, msgs []mensagem) int {
	ultimo := -1
	for _, m := range msgs {
		dados, _ := json.Marshal(m)
		fmt.Fprintf(rw, "id: %d\ndata: %s\n\n", m.ID, dados)
		ultimo = m.ID
	}
	return ultimo
}

func (w *Web) eventos(rw http.ResponseWriter, r *http.Request) {
	if !w.autorizado(r) {
		http.Error(rw, Tr("Acesso negado.", "Access denied."), http.StatusForbidden)
		return
	}
	rw.Header().Set("Content-Type", "text/event-stream")
	controle := http.NewResponseController(rw)

	proximo := 0
	if ultimo, err := strconv.Atoi(r.Header.Get("Last-Event-ID")); err == nil {
		proximo = ultimo + 1
	}

	for {
		msgs, novidade := w.desde(proximo)
		if ultimo := escreverEventos(rw, msgs); ultimo >= 0 {
			proximo = ultimo + 1
		}
		if controle.Flush() != nil {
			return
		}

		select {
		case <-novidade:
		case <-r.Context().Done():
			return
		case <-w.encerrada:
			msgs, _ = w.desde(proximo)
			escreverEventos(rw, msgs)
			controle.Flush()
			return
		}
	}
}

func (w *Web) responder(rw http.ResponseWriter, r *http.Request) {
	if !w.autorizado(r) {
		http.Error(rw, Tr("Acesso negado.", "Access denied."), http.StatusForbidden)
		return
	}
	var corpo struct {
		Resposta string `json:"resposta"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(rw, r.Body, 4096)).Decode(&corpo); err != nil {
		http.Error(rw, Tr("Requisição inválida.", "Invalid request."), http.StatusBadRequest)
		return
	}

	w.mu.Lock()
	aguardando := w.aguardando
	w.aguardando = false
	w.mu.Unlock()
	if !aguardando {
		http.Error(rw, Tr("Nenhuma pergunta aguardando resposta.", "No question is waiting for an answer."), http.StatusConflict)
		return
	}

	w.emitir("resposta", corpo.Resposta)
	w.respostas <- corpo.Resposta
	rw.WriteHeader(http.StatusNoContent)
}
