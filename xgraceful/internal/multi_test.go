package internal

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// blocker simulates a server whose Start blocks until Shutdown is called
// (like http.Server.ListenAndServe).
type blocker struct {
	name   string
	done   chan struct{}
	once   sync.Once
	record *shutdownLog
}

func newBlocker(name string, record *shutdownLog) *blocker {
	return &blocker{name: name, done: make(chan struct{}), record: record}
}

func (b *blocker) Start(ctx context.Context) error {
	<-b.done
	return nil
}

func (b *blocker) Shutdown(ctx context.Context) error {
	b.once.Do(func() {
		if b.record != nil {
			b.record.add(b.name)
		}
		close(b.done)
	})
	return nil
}

// failer simulates a server that fails to start (e.g. port already in use).
type failer struct{ err error }

func (f *failer) Start(ctx context.Context) error    { return f.err }
func (f *failer) Shutdown(ctx context.Context) error { return nil }

type shutdownLog struct {
	mu    sync.Mutex
	order []string
}

func (l *shutdownLog) add(name string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.order = append(l.order, name)
}

// When one service fails to start, Start must return that error rather than
// hanging on the still-running peers.
func TestStart_FailingServiceDoesNotHang(t *testing.T) {
	wantErr := errors.New("bind: address already in use")
	m := NewMulti(newBlocker("a", nil), &failer{err: wantErr}, newBlocker("c", nil))

	errCh := make(chan error, 1)
	go func() { errCh <- m.Start(context.Background()) }()

	select {
	case err := <-errCh:
		require.Error(t, err)
		assert.ErrorIs(t, err, wantErr)
	case <-time.After(2 * time.Second):
		t.Fatal("Start hung after a sub-service failed")
	}
}

// Services must be shut down in reverse registration order so request-serving
// services stop before the telemetry they depend on.
func TestShutdown_ReverseOrder(t *testing.T) {
	log := &shutdownLog{}
	m := NewMulti(newBlocker("first", log), newBlocker("second", log), newBlocker("third", log))

	go func() { _ = m.Start(context.Background()) }()
	time.Sleep(50 * time.Millisecond) // let Start spin up

	require.NoError(t, m.Shutdown(context.Background()))
	assert.Equal(t, []string{"third", "second", "first"}, log.order)
}

// Shutdown is idempotent: calling it again does not re-run teardown.
func TestShutdown_Idempotent(t *testing.T) {
	log := &shutdownLog{}
	m := NewMulti(newBlocker("a", log), newBlocker("b", log))

	require.NoError(t, m.Shutdown(context.Background()))
	require.NoError(t, m.Shutdown(context.Background()))
	assert.Equal(t, []string{"b", "a"}, log.order)
}
