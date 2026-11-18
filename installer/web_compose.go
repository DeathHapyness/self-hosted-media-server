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
		return fmt.Errorf(Tr("não foi possível preparar a validação: %w", "could not prepare the validation: %w"), err)
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
		http.Error(rw, Tr("Serviço não encontrado.", "Service not found."), http.StatusNotFound)
		return
	}
	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		http.Error(rw, Tr("Não foi possível ler o arquivo.", "Could not read the file."), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.Write(conteudo)
}

func (w *Web) validarCompose(rw http.ResponseWriter, r *http.Request) {
	caminho, ok := caminhoCompose(r.PathValue("nome"))
	if !ok {
		http.Error(rw, Tr("Serviço não encontrado.", "Service not found."), http.StatusNotFound)
		return
	}
	conteudo, err := lerCorpoCompose(rw, r)
	if err != nil {
		http.Error(rw, Tr("Arquivo grande demais.", "File too large."), http.StatusRequestEntityTooLarge)
		return
	}
	if err := validarConteudoCompose(caminho, conteudo); err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, err.Error()})
		return
	}
	responderJSON(rw, http.StatusOK, resultadoCompose{true, Tr("Compose válido.", "Compose is valid.")})
}

func (w *Web) salvarCompose(rw http.ResponseWriter, r *http.Request) {
	nome := r.PathValue("nome")
	caminho, ok := caminhoCompose(nome)
	if !ok {
		http.Error(rw, Tr("Serviço não encontrado.", "Service not found."), http.StatusNotFound)
		return
	}
	conteudo, err := lerCorpoCompose(rw, r)
	if err != nil {
		http.Error(rw, Tr("Arquivo grande demais.", "File too large."), http.StatusRequestEntityTooLarge)
		return
	}
	if err := validarConteudoCompose(caminho, conteudo); err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, err.Error()})
		return
	}

	if DRY_RUN {
		responderJSON(rw, http.StatusOK, resultadoCompose{true, Tr("[dry-run] Compose válido; nada foi gravado.", "[dry-run] Compose is valid; nothing was written.")})
		return
	}

	original, err := os.ReadFile(caminho)
	if err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, Tr("Não foi possível ler o arquivo atual.", "Could not read the current file.")})
		return
	}
	info, _ := os.Stat(caminho)
	if err := os.WriteFile(caminho+".bak", original, info.Mode().Perm()); err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, Tr("Não foi possível criar o backup .bak.", "Could not create the .bak backup.")})
		return
	}
	if err := os.WriteFile(caminho, conteudo, info.Mode().Perm()); err != nil {
		responderJSON(rw, http.StatusOK, resultadoCompose{false, Tr("Não foi possível gravar o arquivo.", "Could not write the file.")})
		return
	}

	slog.Warn("Compose edited through the web interface", "service", nome, "path", caminho)
	Registrar("compose_edited", "service", nome, "backup", caminho+".bak")
	w.emitir("ok", fmt.Sprintf(Tr("%s: docker-compose salvo (backup em %s.bak)", "%s: docker-compose saved (backup in %s.bak)"), nome, filepath.Base(caminho)))
	responderJSON(rw, http.StatusOK, resultadoCompose{true, fmt.Sprintf(Tr("Salvo. O original ficou em %s.bak.", "Saved. The original was kept in %s.bak."), filepath.Base(caminho))})
}
