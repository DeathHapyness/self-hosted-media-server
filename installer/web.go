package installer

import (
	"errors"
	"regexp"
	"strings"
	"sync"
)

var codigosANSI = regexp.MustCompile(`\x1b\[[0-9;]*m`)

var errWebEncerrada = errors.New("interface web encerrada")

type mensagem struct {
	ID    int    `json:"id"`
	Tipo  string `json:"tipo"`
	Texto string `json:"texto"`
}

type Web struct {
	mu         sync.Mutex
	mensagens  []mensagem
	novidade   chan struct{}
	aguardando bool

	respostas chan string
	encerrada chan struct{}
	encerrar  sync.Once
}

func novaWeb() *Web {
	return &Web{
		novidade:  make(chan struct{}),
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
