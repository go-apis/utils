package gcp

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-apis/utils/xstorage"
)

const CONCURRENT_SIZE_REQUESTS = 32

type fileStorage struct {
	bucket  string
	service GCSAPI
}

func (store *fileStorage) WriteChunk(ctx context.Context, namespace string, key string, offset int64, src io.Reader) (int64, error) {
	cid := fmt.Sprintf("%s_%d", store.GetPath(ctx, namespace, key), offset)
	objectParams := GCSObjectParams{
		Bucket: store.bucket,
		ID:     cid,
	}

	n, err := store.service.WriteObject(ctx, objectParams, src)
	if err != nil {
		return 0, err
	}

	return n, err
}

func (store *fileStorage) GetReader(ctx context.Context, namespace string, key string) (io.ReadCloser, error) {
	params := GCSObjectParams{
		Bucket: store.bucket,
		ID:     store.GetPath(ctx, namespace, key),
	}

	r, err := store.service.ReadObject(ctx, params)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (store *fileStorage) GetMetadata(ctx context.Context, namespace string, key string) (map[string]string, error) {
	params := GCSObjectParams{
		Bucket: store.bucket,
		ID:     store.GetPath(ctx, namespace, key),
	}

	return store.service.GetObjectMetadata(ctx, params)
}

func (store *fileStorage) FinishUpload(ctx context.Context, namespace string, key string, metadata map[string]string) error {
	p := store.GetPath(ctx, namespace, key)
	prefix := fmt.Sprintf("%s_", p)
	filterParams := GCSFilterParams{
		Bucket: store.bucket,
		Prefix: prefix,
	}

	names, err := store.service.FilterObjects(ctx, filterParams)
	if err != nil {
		return err
	}

	composeParams := GCSComposeParams{
		Bucket:      store.bucket,
		Destination: p,
		Sources:     names,
	}

	err = store.service.ComposeObjects(ctx, composeParams)
	if err != nil {
		return err
	}

	err = store.service.DeleteObjectsWithFilter(ctx, filterParams)
	if err != nil {
		return err
	}

	objectParams := GCSObjectParams{
		Bucket: store.bucket,
		ID:     p,
	}

	err = store.service.SetObjectMetadata(ctx, objectParams, metadata)
	if err != nil {
		return err
	}

	return nil
}

func (store *fileStorage) GetPath(ctx context.Context, namespace string, key string) string {
	var prefix string
	if len(namespace) > 0 {
		prefix = namespace
	}
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	index := strings.Index(key, "/")
	if index > 0 {
		prefix += key[:index+1]
		key = key[index+1:]
	}

	if len(key) > 1 {
		prefix += key[0:2] + "/"
	}
	return prefix + key
}

// New constructs a new GCS storage backend using the supplied GCS bucket name
// and service object.
func NewFileStorage(bucket string, projectId string, service GCSAPI) (xstorage.FileStorage, error) {
	if service != nil {
		ctx := context.Background()
		if err := service.CreateBucket(ctx, GCSBucketParams{
			Bucket:    bucket,
			ProjectId: projectId,
		}); err != nil {
			return nil, err
		}
	}

	return &fileStorage{
		bucket:  bucket,
		service: service,
	}, nil
}
