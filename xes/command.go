package xes

import (
	"context"
	"fmt"
	"reflect"

	"github.com/contextcloud/eventstore/es"
	"github.com/google/uuid"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

type Return struct {
	Id uuid.UUID `json:"id" format:"uuid" required:"true"`
}

func NewCommandInteractor[T es.Command]() usecase.Interactor {
	var cmd T
	u := usecase.NewIOI(cmd, new(Return), func(ctx context.Context, input interface{}, output interface{}) error {
		var (
			in  = input.(T)
			out = output.(*Return)
		)

		// do it!
		unit, err := es.GetUnit(ctx)
		if err != nil {
			return err
		}

		if err := unit.Dispatch(ctx, in); err != nil {
			return err
		}

		out.Id = in.GetAggregateId()
		return nil
	})

	t := reflect.TypeOf(cmd)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	u.SetTitle(fmt.Sprintf("Command %s", t.Name()))
	u.SetName(fmt.Sprintf("Command.%s", t.Name()))
	u.SetExpectedErrors(status.InvalidArgument)
	return u
}
