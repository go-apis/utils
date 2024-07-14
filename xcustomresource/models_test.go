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
