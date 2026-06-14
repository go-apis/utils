package folder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-apis/utils/xstorage"
)

var defaultFilePerm = os.FileMode(0664)

type fileStorage struct {
	abs string
}

// resolve joins the namespace/key under the storage root and verifies the
// result stays within it, rejecting traversal attempts (e.g. "../") before
// any filesystem access happens.
func (store *fileStorage) resolve(ctx context.Context, namespace string, key string) (string, error) {
	rel := store.GetPath(ctx, namespace, key)
	p := filepath.Join(store.abs, rel)
	if p != store.abs && !strings.HasPrefix(p, store.abs+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path: %q escapes storage root", rel)
	}
	return p, nil
}

func (store *fileStorage) WriteChunk(ctx context.Context, namespace string, key string, offset int64, src io.Reader) (int64, error) {
	l := fmt.Sprintf("%s_%d", key, offset)
	p, err := store.resolve(ctx, namespace, l)
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(filepath.Dir(p), os.ModePerm); err != nil {
		return 0, err
	}

	file, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, defaultFilePerm)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	n, err := io.Copy(file, src)
	return n, err
}
func (store *fileStorage) GetReader(ctx context.Context, namespace string, key string) (io.ReadCloser, error) {
	p, err := store.resolve(ctx, namespace, key)
	if err != nil {
		return nil, err
	}
	return os.Open(p)
}
func (store *fileStorage) GetMetadata(ctx context.Context, namespace string, key string) (map[string]string, error) {
	meta := map[string]string{}
	base, err := store.resolve(ctx, namespace, key)
	if err != nil {
		return meta, err
	}
	p := base + ".meta"

	f, err := os.OpenFile(p, os.O_RDONLY, defaultFilePerm)
	if os.IsNotExist(err) {
		return meta, nil
	}
	if err != nil {
		return meta, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&meta); err != nil {
		return meta, err
	}
	return meta, nil
}
func (store *fileStorage) FinishUpload(ctx context.Context, namespace string, key string, metadata map[string]string) error {
	p, err := store.resolve(ctx, namespace, key)
	if err != nil {
		return err
	}
	prefix := fmt.Sprintf("%s_", p)
	var chunks []string

	// walk it?
	if err := filepath.Walk(store.abs, func(p string, info fs.FileInfo, err error) error {
		if info.IsDir() || !strings.HasPrefix(p, prefix) {
			return nil
		}
		chunks = append(chunks, p)
		return nil
	}); err != nil {
		return err
	}

	if len(chunks) == 0 {
		return nil
	}

	file, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, defaultFilePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	sort.Strings(chunks)

	for _, p := range chunks {
		// close each chunk before moving to the next so file descriptors
		// don't accumulate across a many-chunk upload.
		if err := func() error {
			chunk, err := os.OpenFile(p, os.O_RDONLY, defaultFilePerm)
			if err != nil {
				return err
			}
			defer chunk.Close()

			_, err = io.Copy(file, chunk)
			return err
		}(); err != nil {
			return err
		}
	}

	for _, p := range chunks {
		if err := os.Remove(p); err != nil {
			return err
		}
	}

	// save metadata!
	if metadata != nil {
		metafilename := p + ".meta"

		// try truncate!
		if err := os.Truncate(metafilename, 0); err != nil && !os.IsNotExist(err) {
			return err
		}

		m, err := os.OpenFile(metafilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, defaultFilePerm)
		if err != nil {
			return err
		}
		defer m.Close()

		if err := json.NewEncoder(m).Encode(metadata); err != nil {
			return err
		}
	}
	return nil
}
func (store *fileStorage) GetPath(ctx context.Context, namespace string, key string) string {
	if len(namespace) == 0 {
		return key
	}
	return path.Join(namespace, key)
}

func NewFileStorage(p string, remake bool) (xstorage.FileStorage, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}

	if remake {
		os.RemoveAll(abs)
	}

	return &fileStorage{
		abs: abs,
	}, nil
}
