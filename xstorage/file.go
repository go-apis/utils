package xstorage

import (
	"context"
	"io"
)

type FileStorage interface {
	GetPath(ctx context.Context, namespace string, key string) string
	WriteChunk(ctx context.Context, namespace string, key string, offset int64, src io.Reader) (int64, error)
	GetReader(ctx context.Context, namespace string, key string) (io.ReadCloser, error)
	GetMetadata(ctx context.Context, namespace string, key string) (map[string]string, error)
	FinishUpload(ctx context.Context, namespace string, key string, metadata map[string]string) error
}
