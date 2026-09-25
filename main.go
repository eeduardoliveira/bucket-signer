package main

import (
	"bucket-signer/app/usecase"
	"bucket-signer/dependencies/bucket"
	httppresentation "bucket-signer/presentation/http"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	// Swagger docs
	_ "bucket-signer/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Bucket Signer Service
// @version 1.0
// @description API para geração de URLs assinadas para acesso a arquivos no bucket.
// @contact.name SypherTech Team
// @contact.email suporte@syphertech.com.br
// @host bucket-signer.syphertech.com.br
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func main() {
	// Carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Aviso: .env não encontrado, usando variáveis do ambiente")
	}

	s3Presigner, err := bucket.NewS3Presigner()
	if err != nil {
		log.Fatalf("Erro ao inicializar S3Presigner: %v", err)
	}

	var allowed []string
	if raw := strings.TrimSpace(os.Getenv("ALLOWED_BUCKETS")); raw != "" {
		allowed = strings.Split(raw, ",")
	}
	useCase := usecase.NewGenerateURLUseCase(s3Presigner,
		usecase.WithAllowedBuckets(allowed),
		usecase.WithDefaultBucket(os.Getenv("DEFAULT_BUCKET")),
	)
	if !useCase.RestrictsBuckets() {
		log.Println("⚠️⚠️  SEGURANÇA: ALLOWED_BUCKETS não definido — qualquer bucket acessível pelas credenciais AWS pode ser assinado. Defina ALLOWED_BUCKETS=bucket1,bucket2.")
	}

	apiToken := os.Getenv("SIGNER_API_TOKEN")
	if apiToken == "" {
		log.Println("⚠️⚠️  SEGURANÇA: SIGNER_API_TOKEN não definido — /signed-url está SEM autenticação (qualquer um pode gerar URLs de leitura/escrita). Defina SIGNER_API_TOKEN e envie X-API-Key ou Authorization: Bearer.")
	}

	mux := http.NewServeMux()
	httppresentation.RegisterRoutes(mux, &httppresentation.SignedURLController{UseCase: useCase}, apiToken)

	// Swagger UI só quando explicitamente habilitado (desligado por padrão em produção).
	if os.Getenv("ENABLE_SWAGGER") == "true" {
		mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
		log.Println("Swagger habilitado em /swagger/")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	log.Printf("🚀 Servidor iniciado na porta %s", port)
	log.Fatal(srv.ListenAndServe())
}
