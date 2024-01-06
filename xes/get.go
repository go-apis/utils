package xes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/goutils/xlog"
	"github.com/google/uuid"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"go.uber.org/zap"
)

type GetInput struct {
	Namespace string    `header:"X-Namespace" format:"uuid" required:"true"`
	Id        uuid.UUID `path:"id" format:"uuid" required:"true"`
}

func NewGetEntityInteractor[T es.Entity]() usecase.Interactor {
	var entity T
	t := reflect.TypeOf(entity)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	item := reflect.New(t).Interface()

	var in GetInput
	u := usecase.NewIOI(in, item, func(ctx context.Context, input interface{}, output interface{}) error {
		log := xlog.Logger(ctx)

		in := input.(GetInput)

		namespace := es.GetNamespace(ctx)
		if len(in.Namespace) > 0 {
			namespace = in.Namespace
		}
		if in.Id == uuid.Nil {
			return status.InvalidArgument
		}

		unit, err := es.GetUnit(ctx)
		if err != nil {
			return err
		}

		errGet := unit.Get(ctx, name, namespace, in.Id, output)
		if errGet != nil && errors.Is(errGet, sql.ErrNoRows) {
			return status.NotFound
		}
		if errGet != nil {
			log.Error("failed to find config", zap.Error(errGet))
			return fmt.Errorf("failed to find config: %w %w", errGet, status.Unknown)
		}

		return nil
	})

	u.SetTitle(fmt.Sprintf("Get Entity %s", name))
	u.SetName(fmt.Sprintf("Get Entity.%s", name))
	u.SetExpectedErrors(status.InvalidArgument)
	u.SetExpectedErrors(status.Unknown)
	u.SetExpectedErrors(status.NotFound)
	return u
}
