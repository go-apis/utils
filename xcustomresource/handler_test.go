package xcustomresource

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/stretchr/testify/assert"
)

var testCfnEvent = &Event{
	RequestType:       RequestCreate,
	RequestId:         "unique id for this create request",
	ResponseURL:       "http://pre-signed-S3-url-for-response",
	LogicalResourceId: "MyTestResource",
	StackId:           "arn:aws:cloudformation:us-west-2:EXAMPLE/stack-name/guid",
}

var testTfEvent = &Event{
	RequestType: RequestCreate,
	RequestId:   "unique id for this create request",
	TerraformLifecycleScope: &TerraformLifecycleScope{
		Action: TerraformCreate,
	},
}

func TestCopyLambdaLogStream(t *testing.T) {
	lgs := lambdacontext.LogStreamName
	lambdacontext.LogStreamName = "DUMMYLOGSTREAMNAME"

	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			response := extractResponseBody(t, req)

			assert.Equal(t, StatusSuccess, response.Status)
			assert.Equal(t, testCfnEvent.LogicalResourceId, response.LogicalResourceId)
			assert.Equal(t, "DUMMYLOGSTREAMNAME", response.PhysicalResourceId)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nopCloser{bytes.NewBufferString("")},
			}, nil
		},
	}

	h := testHandler(func(ctx context.Context, event Event) (*Success, error) {
		return &Success{
			PhysicalResourceId: event.PhysicalResourceId,
		}, nil
	})

	_, err := h.invokeWithClient(context.TODO(), testCfnEvent, client)
	assert.NoError(t, err)
	lambdacontext.LogStreamName = lgs
}

func TestPanicSendsFailure(t *testing.T) {
	didSendStatus := false

	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			response := extractResponseBody(t, req)
			assert.Equal(t, StatusFailed, response.Status)
			didSendStatus = response.Status == StatusFailed

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nopCloser{bytes.NewBufferString("")},
			}, nil
		},
	}

	h := testHandler(func(ctx context.Context, event Event) (success *Success, err error) {
		err = errors.New("some panic that shouldn't be caught")
		panic(err)
	})
	assert.Panics(t, func() {
		_, err := h.invokeWithClient(context.TODO(), testCfnEvent, client)
		assert.NoError(t, err)
	})

	assert.True(t, didSendStatus, "FAILED should be sent to CloudFormation service")
}

func TestDontCopyLogicalResourceId(t *testing.T) {
	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			response := extractResponseBody(t, req)

			assert.Equal(t, StatusSuccess, response.Status)
			assert.Equal(t, testCfnEvent.LogicalResourceId, response.LogicalResourceId)
			assert.Equal(t, "testingtesting", response.PhysicalResourceId)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nopCloser{bytes.NewBufferString("")},
			}, nil
		},
	}

	h := testHandler(func(ctx context.Context, event Event) (*Success, error) {
		return &Success{
			PhysicalResourceId: "testingtesting",
		}, nil
	})

	_, err := h.invokeWithClient(context.TODO(), testCfnEvent, client)
	assert.NoError(t, err)
}

func TestWrappedError(t *testing.T) {
	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			response := extractResponseBody(t, req)

			assert.Equal(t, StatusFailed, response.Status)
			assert.Empty(t, response.PhysicalResourceId)
			assert.Equal(t, "failed to create resource", response.Reason)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nopCloser{bytes.NewBufferString("")},
			}, nil
		},
	}

	h := testHandler(func(ctx context.Context, event Event) (*Success, error) {
		return nil, errors.New("failed to create resource")
	})
	_, err := h.invokeWithClient(context.TODO(), testCfnEvent, client)
	assert.NoError(t, err)
}

func TestWrappedSendFailure(t *testing.T) {
	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
			}, errors.New("things went wrong")
		},
	}

	h := testHandler(func(ctx context.Context, event Event) (success *Success, e error) {
		return
	})

	_, e := h.invokeWithClient(context.TODO(), testCfnEvent, client)
	assert.NotNil(t, e)
	// _ (r) is nil because the send error is logged only
	assert.Equal(t, "things went wrong", e.Error())
}

func TestEventWithoutResponseURL(t *testing.T) {
	successResponse := Success{
		PhysicalResourceId: "testingtesting",
		Data:               "some data",
	}
	h := testHandler(func(ctx context.Context, event Event) (success *Success, e error) {
		return &successResponse, nil
	})
	payload, e := json.Marshal(testTfEvent)
	assert.NoError(t, e)

	r, e := h.Invoke(context.TODO(), payload)
	assert.NoError(t, e)
	assert.NotNil(t, r)
	// validate success was copied over to response
	var response Response
	json.Unmarshal(r, &response)
	assert.Equal(t, successResponse.PhysicalResourceId, response.PhysicalResourceId)
	assert.Equal(t, successResponse.Data, response.Data)
}

func testHandler(run Run) *handler {
	return &handler{
		run: run,
	}
}

func extractResponseBody(t *testing.T, req *http.Request) Response {
	assert.NotContains(t, req.Header, "Content-Type")

	body, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	var response Response
	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)
	return response
}
