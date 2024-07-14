package xcustomresource

import "encoding/json"

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
	// Terraform lifecycle fields
	// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_invocation#crud-lifecycle-scope
	TerraformLifecycleScope *TerraformLifecycleScope `json:"tf,omitempty"`
}

// action enum Values are create, update, or delete.
type Action string

const (
	TerraformCreate Action = "create"
	TerraformUpdate Action = "update"
	TerraformDelete Action = "delete"
)

type TerraformLifecycleScope struct {
	Action    Action           `json:"action"`
	PrevInput *json.RawMessage `json:"prev_input,omitempty"`
}
