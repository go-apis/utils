package xes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/go-apis/utils/xlog"
	"github.com/google/uuid"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"go.uber.org/zap"
)

type GetInput struct {
	Namespace string    `header:"X-Namespace" required:"true"`
	Id        uuid.UUID `path:"id" format:"uuid" required:"true"`
}

func NewGetEntityInteractor[T es.Entity]() usecase.Interactor {
	var entity T
	opts := es.NewEntityOptions(entity)
	entityConfig, err := es.NewEntityConfig(opts)
	if err != nil {
		panic(err)
	}
	item, err := entityConfig.Factory()
	if err != nil {
		panic(err)
	}

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

		errGet := unit.Get(ctx, entityConfig.Name, namespace, in.Id, output)
		if errGet != nil && errors.Is(errGet, sql.ErrNoRows) {
			return status.NotFound
		}
		if errGet != nil {
			log.Error("failed to find", zap.String("name", entityConfig.Name), zap.Error(errGet))
			return fmt.Errorf("failed to find config: %w %w", errGet, status.Unknown)
		}

		return nil
	})

	u.SetTitle(fmt.Sprintf("Get %s", entityConfig.Name))
	u.SetName(fmt.Sprintf("Get %s", entityConfig.Name))
	u.SetExpectedErrors(status.InvalidArgument)
	u.SetExpectedErrors(status.Unknown)
	u.SetExpectedErrors(status.NotFound)
	return u
}
