package xcustomresource

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events/test"
	"github.com/stretchr/testify/assert"
)

func TestCloudFormationEventMarshaling(t *testing.T) {
	// read json from file
	inputJSON := test.ReadJSONFromFile(t, "./testdata/cloudformation-event.json")
	// de-serialize into Event
	var inputEvent Event
	if err := json.Unmarshal(inputJSON, &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}

	// serialize to json
	outputJSON, err := json.Marshal(inputEvent)
	if err != nil {
		t.Errorf("could not marshal event. details: %v", err)
	}

	test.AssertJsonsEqual(t, inputJSON, outputJSON)
}

func TestCloudFormationMarshalingMalformedJson(t *testing.T) {
	test.TestMalformedJson(t, Event{})
}

func TestTerraformCreateEventMarshaling(t *testing.T) {
	inputJSON := test.ReadJSONFromFile(t, "./testdata/terraform-event-create.json")
	// de-serialize into Event
	var inputEvent Event
	if err := json.Unmarshal(inputJSON, &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}

	assert.NotNil(t, inputEvent.TerraformLifecycleScope)
	assert.Equal(t, TerraformCreate, inputEvent.TerraformLifecycleScope.Action)
	assert.Nil(t, inputEvent.TerraformLifecycleScope.PrevInput)
}

func TestTerraformUpdateEventMarshaling(t *testing.T) {
	inputJSON := test.ReadJSONFromFile(t, "./testdata/terraform-event-update.json")
	// de-serialize into Event
	var inputEvent Event
	if err := json.Unmarshal(inputJSON, &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}

	assert.NotNil(t, inputEvent.TerraformLifecycleScope)
	assert.Equal(t, TerraformUpdate, inputEvent.TerraformLifecycleScope.Action)
	assert.NotNil(t, inputEvent.TerraformLifecycleScope.PrevInput)
}

func TestEventActions(t *testing.T) {
	tests := []struct {
		Name           string
		RequestType    RequestType
		Action         *Action
		ExpectedCreate bool
		ExpectedUpdate bool
		ExpectedDelete bool
	}{
		{
			Name:           "Create Event",
			RequestType:    RequestCreate,
			Action:         nil,
			ExpectedCreate: true,
			ExpectedUpdate: false,
			ExpectedDelete: false,
		},
		{
			Name:           "Update Event",
			RequestType:    RequestUpdate,
			Action:         nil,
			ExpectedCreate: false,
			ExpectedUpdate: true,
			ExpectedDelete: false,
		},
		{
			Name:           "Delete Event",
			RequestType:    RequestDelete,
			Action:         nil,
			ExpectedCreate: false,
			ExpectedUpdate: false,
			ExpectedDelete: true,
		},
		{
			Name:           "Terraform Create Event",
			RequestType:    "",
			Action:         ptr(TerraformCreate),
			ExpectedCreate: true,
			ExpectedUpdate: false,
			ExpectedDelete: false,
		},
		{
			Name:           "Terraform Update Event",
			RequestType:    RequestCreate,
			Action:         ptr(TerraformUpdate),
			ExpectedCreate: false,
			ExpectedUpdate: true,
			ExpectedDelete: false,
		},
		{
			Name:           "Terraform Delete Event",
			RequestType:    RequestCreate,
			Action:         ptr(TerraformDelete),
			ExpectedCreate: false,
			ExpectedUpdate: false,
			ExpectedDelete: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			event := Event{
				RequestType: test.RequestType,
			}
			if test.Action != nil {
				event.TerraformLifecycleScope = &TerraformLifecycleScope{
					Action:    *test.Action,
					PrevInput: nil,
				}
			}

			assert.Equal(t, test.ExpectedCreate, event.IsCreate())
			assert.Equal(t, test.ExpectedUpdate, event.IsUpdate())
			assert.Equal(t, test.ExpectedDelete, event.IsDelete())
		})
	}
}

func ptr[T any](s T) *T {
	return &s
}
