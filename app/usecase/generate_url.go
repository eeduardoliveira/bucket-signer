package usecase

import (
	"bucket-signer/app/domain"
	"context"
)

// GenerateURLUseCase representa o caso de uso de geração de URL assinada
type GenerateURLUseCase struct {
	presigner domain.Presigner
}

// NewGenerateURLUseCase cria uma nova instância do caso de uso
func NewGenerateURLUseCase(presigner domain.Presigner) *GenerateURLUseCase {
	return &GenerateURLUseCase{
		presigner: presigner,
	}
}

// Execute executa o caso de uso para o bucket e cliente informado
func (uc *GenerateURLUseCase) Execute(ctx context.Context, bucketName, clienteID string, upload bool) (string, error) {
	return uc.presigner.GeneratePresignedURL(ctx, bucketName, clienteID, upload)
}