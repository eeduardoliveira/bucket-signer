package http

import (
	"net/http"
)

// RegisterRoutes configura as rotas HTTP da aplicação
func RegisterRoutes(controller *SignedURLController) {
	http.HandleFunc("/signed-url", controller.HandleSignedURL)
}