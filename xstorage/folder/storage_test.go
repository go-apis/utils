package folder

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {
	ctx := context.Background()
	store, err := NewFileStorage(t.TempDir(), true)
	require.NoError(t, err)

	_, err = store.WriteChunk(ctx, "ns", "file", 0, strings.NewReader("foo"))
	require.NoError(t, err)
	_, err = store.WriteChunk(ctx, "ns", "file", 1, strings.NewReader("bar"))
	require.NoError(t, err)

	require.NoError(t, store.FinishUpload(ctx, "ns", "file", map[string]string{"k": "v"}))

	r, err := store.GetReader(ctx, "ns", "file")
	require.NoError(t, err)
	defer r.Close()

	data, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "foobar", string(data))

	meta, err := store.GetMetadata(ctx, "ns", "file")
	require.NoError(t, err)
	assert.Equal(t, "v", meta["k"])
}

// Keys or namespaces containing traversal segments must be rejected before any
// filesystem access escapes the storage root.
func TestPathTraversalRejected(t *testing.T) {
	ctx := context.Background()
	store, err := NewFileStorage(t.TempDir(), true)
	require.NoError(t, err)

	_, err = store.GetReader(ctx, "", "../../../etc/passwd")
	assert.Error(t, err)

	_, err = store.WriteChunk(ctx, "..", "file", 0, strings.NewReader("x"))
	assert.Error(t, err)

	err = store.FinishUpload(ctx, "../..", "file", nil)
	assert.Error(t, err)
}
