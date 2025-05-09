package main

import (
	"bucket-signer/app/usecase"
	"bucket-signer/dependencies/bucket"
	httppresentation "bucket-signer/presentation/http"
	"log"
	"net/http"
	"os"

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

	// Rota do Swagger
	http.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./docs"))))
	http.HandleFunc("/swagger/index.html", httpSwagger.WrapHandler)

	// Define porta
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🚀 Servidor iniciado em http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}