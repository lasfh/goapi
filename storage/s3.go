package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Options struct {
	Options

	// Endpoint URL base de um serviço compatível com S3.
	Endpoint string

	// Bucket nome do bucket S3.
	Bucket string

	// Region região (ex.: "us-east-1"). Quando vazio, usa "us-east-1" como padrão.
	Region string

	// AccessKeyID chave de acesso.
	AccessKeyID string

	// SecretKey chave secreta.
	SecretKey string

	// UsePathStyle habilita path-style URLs (necessário para MinIO e alguns provedores).
	UsePathStyle bool

	// HTTPClient cliente HTTP customizado.
	HTTPClient config.HTTPClient
}

type s3Storage struct {
	Options

	bucket string
	client *s3.Client
}

func NewS3(ctx context.Context, opts S3Options) (Storage, error) {
	if strings.TrimSpace(opts.Bucket) == "" {
		return nil, fmt.Errorf("%w: Bucket é obrigatório", ErrInvalidOptions)
	}

	opts.Options.ValidExts = normalizeExtensions(opts.Options.ValidExts)

	region := opts.Region
	if region == "" {
		region = "us-east-1"
	}

	optFns := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if opts.HTTPClient != nil {
		optFns = append(optFns, config.WithHTTPClient(opts.HTTPClient))
	}

	optFns = append(optFns, config.WithRegion(region))

	if opts.AccessKeyID != "" && opts.SecretKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.AccessKeyID, opts.SecretKey, ""),
		))
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("s3: carregar configuração: %w", err)
	}

	s3Opts := []func(*s3.Options){
		func(o *s3.Options) {
			o.UsePathStyle = opts.UsePathStyle
		},
	}

	if opts.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(opts.Endpoint)
		})
	}

	return &s3Storage{
		Options: opts.Options,
		bucket:  opts.Bucket,
		client:  s3.NewFromConfig(cfg, s3Opts...),
	}, nil
}

func (s *s3Storage) WithExtValidation(exts ...string) Storage {
	c := *s
	c.Options.ValidExts = normalizeExtensions(exts)

	return &c
}

func (s *s3Storage) Read(ctx context.Context, path string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(path)),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get %s: %w", path, err)
	}

	return out.Body, nil
}

func (s *s3Storage) ReadAll(ctx context.Context, path string) ([]byte, error) {
	rc, err := s.Read(ctx, path)
	if err != nil {
		return nil, err
	}

	defer rc.Close()

	return io.ReadAll(rc)
}

func (s *s3Storage) Create(ctx context.Context, path string, r io.Reader) error {
	if err := s.checkExt(path); err != nil {
		return err
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(path)),
		Body:   r,
	})
	if err != nil {
		return fmt.Errorf("s3 put %s: %w", path, err)
	}

	return nil
}

func (s *s3Storage) Remove(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(path)),
	})
	if err != nil {
		return fmt.Errorf("s3 delete %s: %w", path, err)
	}

	return nil
}

func (s *s3Storage) key(path string) string {
	return strings.TrimPrefix(path, "/")
}
