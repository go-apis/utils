package xcustomresource

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataCopiedFromRequest(t *testing.T) {
	e := Event{
		RequestId:         "unique id for this create request",
		ResponseURL:       "http://pre-signed-S3-url-for-response",
		LogicalResourceId: "MyTestResource",
		StackId:           "arn:aws:cloudformation:us-west-2:EXAMPLE/stack-name/guid",
	}
	success := &Success{
		PhysicalResourceId: "MyTestResourceId",
	}

	r := createResponse(context.Background(), e, success, nil)
	assert.Equal(t, e.RequestId, r.RequestId)
	assert.Equal(t, e.LogicalResourceId, r.LogicalResourceId)
	assert.Equal(t, e.StackId, r.StackId)
	assert.Equal(t, e.ResponseURL, r.url)
	assert.Equal(t, success.PhysicalResourceId, r.PhysicalResourceId)
	assert.Equal(t, success.Data, r.Data)
}

type mockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{}, nil
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestRequestSentCorrectly(t *testing.T) {
	r := &Response{
		Status: StatusSuccess,
		url:    "http://pre-signed-S3-url-for-response",
	}

	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.NotContains(t, req.Header, "Content-Type")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nopCloser{bytes.NewBufferString("")},
			}, nil
		},
	}

	assert.NoError(t, r.sendWith(context.Background(), client))
}

func TestRequestForbidden(t *testing.T) {
	r := &Response{
		Status: StatusSuccess,
		url:    "http://pre-signed-S3-url-for-response",
	}

	sc := http.StatusForbidden
	client := &mockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.NotContains(t, req.Header, "Content-Type")
			return &http.Response{
				StatusCode: sc,
				Body:       nopCloser{bytes.NewBufferString("")},
			}, nil
		},
	}

	s := r.sendWith(context.Background(), client)
	if assert.Error(t, s) {
		assert.Equal(t, fmt.Errorf("invalid status code. got: %d", sc), s)
	}
}
