package bucket

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Presigner struct {
	client     *s3.Client
	expiration time.Duration
}

func NewS3Presigner() (*S3Presigner, error) {
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("variáveis AWS_ACCESS_KEY_ID ou AWS_SECRET_ACCESS_KEY não definidas")
	}

	// Carrega configuração com credenciais estáticas para AWS S3
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar configuração AWS: %w", err)
	}

	// Cria cliente S3
	s3Client := s3.NewFromConfig(cfg)

	// Define tempo de expiração da URL
	expMin := os.Getenv("EXPIRATION_MINUTES")
	if expMin == "" {
		expMin = "10"
	}
	exp, _ := time.ParseDuration(expMin + "m")

	return &S3Presigner{
		client:     s3Client,
		expiration: exp,
	}, nil
}

// GeneratePresignedURL gera a URL assinada com base no clienteID
func (p *S3Presigner) GeneratePresignedURL(ctx context.Context, bucketName, clienteID string) (string, error) {
	clienteID = strings.TrimSpace(clienteID)
	key := fmt.Sprintf(os.Getenv("PROMPT_FILE_PATTERN"), clienteID, clienteID)

	fmt.Println("🔑 Nome final do objeto no bucket:", key)

	presigner := s3.NewPresignClient(p.client)

	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(p.expiration))

	if err != nil {
		return "", fmt.Errorf("erro ao gerar presigned URL: %w", err)
	}

	return req.URL, nil
}