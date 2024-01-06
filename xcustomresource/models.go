package xcustomresource

import "encoding/json"

// StatusType represents a CloudFormation response status
type StatusType string

const (
	StatusSuccess StatusType = "SUCCESS"
	StatusFailed  StatusType = "FAILED"
)

// RequestType represents a CloudFormation request type
type RequestType string

const (
	RequestCreate RequestType = "Create"
	RequestUpdate RequestType = "Update"
	RequestDelete RequestType = "Delete"
)

// Event represents a CloudFormation request
type Event struct {
	RequestId             string          `json:"RequestId"`
	StackId               string          `json:"StackId"`
	RequestType           RequestType     `json:"RequestType"`
	ResourceType          string          `json:"ResourceType"`
	LogicalResourceId     string          `json:"LogicalResourceId"`
	PhysicalResourceId    string          `json:"PhysicalResourceId,omitempty"`
	ResponseURL           string          `json:"ResponseURL"`
	ResourceProperties    json.RawMessage `json:"ResourceProperties"`
	OldResourceProperties json.RawMessage `json:"OldResourceProperties,omitempty"`
	ServiceToken          string
}

type Success struct {
	PhysicalResourceId string      `json:"PhysicalResourceId"`
	Data               interface{} `json:"Data,omitempty"`
}

type Response struct {
	Status             StatusType  `json:"Status"`
	StackId            string      `json:"StackId"`
	RequestId          string      `json:"RequestId"`
	PhysicalResourceId string      `json:"PhysicalResourceId"`
	LogicalResourceId  string      `json:"LogicalResourceId"`
	Reason             string      `json:"Reason,omitempty"`
	NoEcho             bool        `json:"NoEcho,omitempty"`
	Data               interface{} `json:"Data,omitempty"`
}
