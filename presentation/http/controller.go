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

func (c *SignedURLController) HandleSignedURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	clienteID := r.URL.Query().Get("clienteID")
	bucketName := r.URL.Query().Get("bucket")

	if clienteID == "" || bucketName == "" {
		http.Error(w, `{"error":"bucket e clienteID são obrigatórios"}`, http.StatusBadRequest)
		return
	}

	url, err := c.UseCase.Execute(r.Context(), bucketName, clienteID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"erro ao gerar URL: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	url = html.UnescapeString(url)
	resposta := fmt.Sprintf(`{"url":"%s"}`, url)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resposta))
}