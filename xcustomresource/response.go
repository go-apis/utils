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

// StatusType represents a CloudFormation response status
type StatusType string

const (
	StatusSuccess StatusType = "SUCCESS"
	StatusFailed  StatusType = "FAILED"
)

// Response is a representation of a Custom Resource
// response expected by CloudFormation.
type Response struct {
	Status             StatusType  `json:"Status"`
	StackId            string      `json:"StackId"`
	RequestId          string      `json:"RequestId"`
	PhysicalResourceId string      `json:"PhysicalResourceId"`
	LogicalResourceId  string      `json:"LogicalResourceId"`
	Reason             string      `json:"Reason,omitempty"`
	NoEcho             bool        `json:"NoEcho,omitempty"`
	Data               interface{} `json:"Data,omitempty"`

	url string
}

// NewResponse creates a Response with the relevant verbatim copied
// data from an Event
func NewResponse(e *Event) *Response {
	return &Response{
		RequestId:         e.RequestId,
		LogicalResourceId: e.LogicalResourceId,
		StackId:           e.StackId,

		url: e.ResponseURL,
	}
}

type Success struct {
	PhysicalResourceId string      `json:"PhysicalResourceId"`
	Data               interface{} `json:"Data,omitempty"`
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (r *Response) sendWith(ctx context.Context, client httpClient) error {
	log := xlog.Logger(ctx)
	body, err := json.Marshal(r)
	if err != nil {
		log.Error("failed to marshal response body", zap.Error(err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, r.url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Del("Content-Type")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Error("failed to send response", zap.Int("status", res.StatusCode), zap.String("body", string(body)))
		return fmt.Errorf("invalid status code. got: %d", res.StatusCode)
	}

	return nil
}

func (r *Response) Send(ctx context.Context) error {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	return r.sendWith(ctx, client)
}

// createResponse creates a Response taking into account success data
func createResponse(ctx context.Context, in Event, success *Success, err error) *Response {
	rsp := NewResponse(&in)
	if success != nil {
		rsp.PhysicalResourceId = success.PhysicalResourceId
		rsp.Data = success.Data
	}
	if rsp.PhysicalResourceId == "" {
		rsp.PhysicalResourceId = lambdacontext.LogStreamName
	}
	rsp.Status = StatusSuccess

	if err != nil {
		rsp.Status = StatusFailed
		rsp.Reason = err.Error()
	}

	return rsp
}
