package domain

import (
	"context"
)

// Presigner define o contrato para qualquer serviço que consiga gerar uma URL assinada
type Presigner interface {
	// GeneratePresignedURL deve retornar uma URL temporária para acesso ao arquivo do cliente
	GeneratePresignedURL(ctx context.Context, bucketName, clienteID string, upload bool) (string, error)
}