package xes

import (
	"testing"

	"github.com/go-apis/eventsourcing/es"
)

type TestEntity struct {
	es.BaseAggregateHolder

	Something string `json:"something"`
}

type TestWhere struct {
	BaseFindInput
}

func Test_Usecases(t *testing.T) {
	uc1 := NewGetEntityInteractor[*TestEntity]()
	if uc1 == nil {
		t.Fail()
	}

	uc2 := NewFindEntityInteractor[*TestEntity, *TestWhere]()
	if uc2 == nil {
		t.Fail()
	}
}
