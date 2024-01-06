package xcustomresource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/contextcloud/goutils/xlog"
	"go.uber.org/zap"
)

func cloudformationCallback(ctx context.Context, url string, data *Response) error {
	log := xlog.Logger(ctx)

	body, err := json.Marshal(data)
	if err != nil {
		log.Error("failed to marshal response body", zap.Error(err))
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Del("Content-Type")

	client := http.Client{
		Timeout: 15 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Error("failed to send response", zap.Int("status", res.StatusCode), zap.String("body", string(out)))
		return fmt.Errorf("invalid status code. got: %d", res.StatusCode)
	}

	return nil
}

func createResponse(ctx context.Context, in Event, success *Success, err error) *Response {
	physicalId := in.PhysicalResourceId
	var data interface{}

	if success != nil {
		physicalId = success.PhysicalResourceId
		data = success.Data
	}
	if physicalId == "" {
		physicalId = lambdacontext.LogStreamName
	}

	rsp := &Response{
		Status:             StatusSuccess,
		StackId:            in.StackId,
		RequestId:          in.RequestId,
		LogicalResourceId:  in.LogicalResourceId,
		PhysicalResourceId: physicalId,
		Data:               data,
	}

	if err != nil {
		rsp.Status = StatusFailed
		rsp.Reason = err.Error()
	}

	return rsp
}
