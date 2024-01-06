package xcustomresource

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/contextcloud/goutils/xlog"
	"go.uber.org/zap"
)

type Run func(ctx context.Context, event Event) (*Success, error)

type handler struct {
	run Run
}

func (h *handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	log := xlog.Logger(ctx)
	log.Info("payload", zap.String("payload", string(payload)))

	// do stuff
	var in Event
	if err := json.Unmarshal(payload, &in); err != nil {
		log.Error("failed to unmarshal payload", zap.Error(err))
		return nil, err
	}

	// do stuff
	success, err := h.run(ctx, in)
	if err != nil {
		log.Error("failed to run", zap.Error(err))
	}

	rsp := createResponse(ctx, in, success, err)
	if err := cloudformationCallback(ctx, in.ResponseURL, rsp); err != nil {
		log.Error("failed to send response", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func NewHandler(run Run) lambda.Handler {
	return &handler{
		run: run,
	}
}
