package main

import (
	"bucket-signer/app/usecase"
	"bucket-signer/dependencies/bucket"
	httppresentation "bucket-signer/presentation/http"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Carrega variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Aviso: .env não encontrado, usando variáveis do ambiente")
	}

	// Instancia a dependência (S3Presigner)
	s3Presigner, err := bucket.NewS3Presigner()
	if err != nil {
		log.Fatalf("Erro ao inicializar S3Presigner: %v", err)
	}

	// Cria o caso de uso
	useCase := usecase.NewGenerateURLUseCase(s3Presigner)

	// Cria o controller com o use case
	controller := &httppresentation.SignedURLController{
		UseCase: useCase,
	}

	// Registra as rotas HTTP
	httppresentation.RegisterRoutes(controller)

	// Define porta
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🚀 Servidor iniciado em http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}