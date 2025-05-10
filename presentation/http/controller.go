package http

import (
	"bucket-signer/app/usecase"
	"fmt"
	"html"
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
// HandleSignedURL godoc
// @Summary Gera uma URL assinada para acesso ao bucket
// @Description Gera uma URL temporária para upload ou download de um arquivo no bucket
// @Tags bucket
// @Accept json
// @Produce json
// @Param bucket query string true "Nome do bucket"
// @Param clienteID query string true "ID do cliente"
// @Param upload query bool false "Define se a URL será para upload (true) ou download (false)"
// @Success 200 {object} signedURLResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /signed-url [get]
func (c *SignedURLController) HandleSignedURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	clienteID := r.URL.Query().Get("clienteID")
	bucketName := r.URL.Query().Get("bucket")
	upload := r.URL.Query().Get("upload") == "true"

	if clienteID == "" || bucketName == "" {
		http.Error(w, `{"error":"bucket e clienteID são obrigatórios"}`, http.StatusBadRequest)
		return
	}

	url, err := c.UseCase.Execute(r.Context(), bucketName, clienteID, upload)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"erro ao gerar URL: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	url = html.UnescapeString(url)
	resposta := fmt.Sprintf(`{"url":"%s"}`, url)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resposta))
}