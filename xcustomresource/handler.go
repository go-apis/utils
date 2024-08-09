package xcustomresource

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-apis/utils/xlog"
	"go.uber.org/zap"
)

type Run func(ctx context.Context, event Event) (*Success, error)

type handler struct {
	run Run
}

func (h *handler) invokeWithClient(ctx context.Context, in *Event, client httpClient) ([]byte, error) {
	log := xlog.Logger(ctx)
	runDidPanic := true
	defer func() {
		if runDidPanic {
			r := NewResponse(in)
			r.Status = StatusFailed
			r.Reason = "Run panicked, see log stream for details"
			// FIXME: something should be done if an error is returned here
			_ = r.sendWith(ctx, client)
		}
	}()

	success, err := h.run(ctx, *in)
	runDidPanic = false
	if err != nil {
		log.Error("failed to run", zap.Error(err))
	}

	rsp := createResponse(ctx, *in, success, err)
	if err := rsp.sendWith(ctx, client); err != nil {
		log.Error("failed to send response", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func (h *handler) invokeWithoutCallback(ctx context.Context, in *Event) ([]byte, error) {
	log := xlog.Logger(ctx)
	success, err := h.run(ctx, *in)
	if err != nil {
		log.Error("failed to run", zap.Error(err))
	}
	rsp := createResponse(ctx, *in, success, err)
	result, err := json.Marshal(rsp)
	if err != nil {
		log.Error("failed to marshal response", zap.Error(err))
		return nil, err
	}
	return result, nil
}

func (h *handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	log := xlog.Logger(ctx)
	log.Info("payload", zap.String("payload", string(payload)))

	// TODO: unmarshal panics aren't caught
	var in Event
	if err := json.Unmarshal(payload, &in); err != nil {
		log.Error("failed to unmarshal payload", zap.Error(err))
		return nil, err
	}
	if in.ResponseURL == "" {
		log.Info("missing response URL, returning response directly")
		return h.invokeWithoutCallback(ctx, &in)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	return h.invokeWithClient(ctx, &in, client)
}

func NewHandler(run Run) lambda.Handler {
	return &handler{
		run: run,
	}
}
