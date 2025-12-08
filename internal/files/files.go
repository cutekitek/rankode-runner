package files

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileStorage struct {
	cl     *minio.Client
	Bucket string
}

type Config struct {
	Url      string
	Login    string
	Password string
	Bucket   string
}

func NewFileStorage(cfg Config) *FileStorage {
	client, err := minio.New(cfg.Url, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.Login, cfg.Password, ""),
	})
	if err != nil {
		panic(err)
	}
	return &FileStorage{cl: client, Bucket: cfg.Bucket}
}

func (s *FileStorage) GetFile(ctx context.Context, filename string) (io.Reader, error) {
	file, err := s.cl.GetObject(ctx, s.Bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return file, nil
}
