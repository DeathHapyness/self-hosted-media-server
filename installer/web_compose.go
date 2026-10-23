package installer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

const limiteCompose = 1 << 20

type resultadoCompose struct {
	Ok       bool   `json:"ok"`
	Mensagem string `json:"mensagem"`
}

func responderJSON(rw http.ResponseWriter, status int, v any) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(v)
}

func caminhoCompose(nome string) (string, bool) {
	caminho, ok := findComposeFiles()[nome]
	return caminho, ok
}

func validarConteudoCompose(caminho string, conteudo []byte) error {
	temporario := filepath.Join(filepath.Dir(caminho), ".docker-compose.validar.yml")
	if err := os.WriteFile(temporario, conteudo, 0o600); err != nil {
		return fmt.Errorf("não foi possível preparar a validação: %w", err)
	}
	defer os.Remove(temporario)

	_, stderr, err := runCompose("-f", temporario, "--project-directory", filepath.Dir(caminho), "config", "-q")
	if err != nil {
		if stderr == "" {
			stderr = err.Error()
		}
		return errors.New(stderr)
	}
	return nil
}

func lerCorpoCompose(rw http.ResponseWriter, r *http.Request) ([]byte, error) {
	return io.ReadAll(http.MaxBytesReader(rw, r.Body, limiteCompose))
}

func (w *Web) listarCompose(rw http.ResponseWriter, r *http.Request) {
	responderJSON(rw, http.StatusOK, nomesOrdenados(findComposeFiles()))
}

func (w *Web) lerCompose(rw http.ResponseWriter, r *http.Request) {
	caminho, ok := caminhoCompose(r.PathValue("nome"))
	if !ok {
		http.Error(rw, "Serviço não encontrado.", http.StatusNotFound)
		return
	}
	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		http.Error(rw, "Não foi possível ler o arquivo.", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.Write(conteudo)
}

func (w *Web) validarCompose(rw http.ResponseWriter, r *http.Request) {
	caminho, ok := caminhoCompose(r.PathValue("nome"))
	if !ok {
		http.Error(rw, "Serviço não encontrado.", http.StatusNotFound)
		return
	}
	conteudo, err := lerCorpoCompose(rw, r)
	if err != nil {
		http.Error(rw, "Arquivo grande demais.", http.StatusRequestEntityTooLarge)
		return
	}
	if err := validarConteudoCompose(caminho, conteudo); err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, err.Error()})
		return
	}
	responderJSON(rw, http.StatusOK, resultadoCompose{true, "Compose válido."})
}

func (w *Web) salvarCompose(rw http.ResponseWriter, r *http.Request) {
	nome := r.PathValue("nome")
	caminho, ok := caminhoCompose(nome)
	if !ok {
		http.Error(rw, "Serviço não encontrado.", http.StatusNotFound)
		return
	}
	conteudo, err := lerCorpoCompose(rw, r)
	if err != nil {
		http.Error(rw, "Arquivo grande demais.", http.StatusRequestEntityTooLarge)
		return
	}
	if err := validarConteudoCompose(caminho, conteudo); err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, err.Error()})
		return
	}

	if DRY_RUN {
		responderJSON(rw, http.StatusOK, resultadoCompose{true, "[dry-run] Compose válido; nada foi gravado."})
		return
	}

	original, err := os.ReadFile(caminho)
	if err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, "Não foi possível ler o arquivo atual."})
		return
	}
	info, _ := os.Stat(caminho)
	if err := os.WriteFile(caminho+".bak", original, info.Mode().Perm()); err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, "Não foi possível criar o backup .bak."})
		return
	}
	if err := os.WriteFile(caminho, conteudo, info.Mode().Perm()); err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, "Não foi possível gravar o arquivo."})
		return
	}

	slog.Warn("Compose editado pela interface web", "servico", nome, "caminho", caminho)
	Registrar("compose_editado", "servico", nome, "backup", caminho+".bak")
	w.emitir("ok", fmt.Sprintf("%s: docker-compose salvo (backup em %s.bak)", nome, filepath.Base(caminho)))
	responderJSON(rw, http.StatusOK, resultadoCompose{true, "Salvo. O original ficou em " + filepath.Base(caminho) + ".bak."})
}
