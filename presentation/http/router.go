package http

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// RegisterRoutes configura as rotas HTTP da aplicação no mux informado.
// Se apiToken não for vazio, /signed-url exige "Authorization: Bearer <token>" ou "X-API-Key: <token>".
func RegisterRoutes(mux *http.ServeMux, controller *SignedURLController, apiToken string) {
	mux.Handle("/signed-url", RequireToken(apiToken, http.HandlerFunc(controller.HandleSignedURL)))
}

// RequireToken protege next com um token estático. Token vazio = sem autenticação (modo legado).
func RequireToken(token string, next http.Handler) http.Handler {
	if token == "" {
		return next
	}
	want := []byte(token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-API-Key")
		if auth := r.Header.Get("Authorization"); got == "" && len(auth) > 7 && strings.EqualFold(auth[:7], "Bearer ") {
			got = strings.TrimSpace(auth[7:])
		}
		if got == "" || subtle.ConstantTimeCompare([]byte(got), want) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="bucket-signer"`)
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "não autorizado"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
