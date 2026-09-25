package http

import (
	"bucket-signer/app/domain"
	"bucket-signer/app/usecase"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type SignedURLController struct {
	UseCase *usecase.GenerateURLUseCase
}

type signedURLResponse struct {
	URL string `json:"url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// HandleSignedURL godoc
// @Summary Gera uma URL assinada para acesso ao bucket
// @Description Gera uma URL temporária para upload ou download de um arquivo no bucket
// @Tags bucket
// @Accept json
// @Produce json
// @Param bucket query string false "Nome do bucket (default: DEFAULT_BUCKET; restrito a ALLOWED_BUCKETS se configurado)"
// @Param clienteID query string true "ID do cliente ([A-Za-z0-9._@-], sem '/' ou '..')"
// @Param upload query bool false "Define se a URL será para upload (true) ou download (false)"
// @Security ApiKeyAuth
// @Success 200 {object} signedURLResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /signed-url [get]
func (c *SignedURLController) HandleSignedURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "método não permitido"})
		return
	}

	q := r.URL.Query()
	clienteID := q.Get("clienteID")
	if clienteID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "clienteID é obrigatório"})
		return
	}

	url, err := c.UseCase.Execute(r.Context(), q.Get("bucket"), clienteID, q.Get("upload") == "true")
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, signedURLResponse{URL: url})
	case errors.Is(err, domain.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "parâmetros inválidos: " + err.Error()})
	case errors.Is(err, domain.ErrBucketNotAllowed):
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "bucket não permitido"})
	default:
		// Detalhes internos (AWS/SDK) ficam só no log.
		log.Printf("erro ao gerar URL assinada: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "erro ao gerar URL"})
	}
}
