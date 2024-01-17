package xes

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/goutils/xlog"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"go.uber.org/zap"
)

type FindInput[W any] struct {
	Where     W          `query:"where" required:"true"`
	Limit     *int       `query:"limit"`
	Offset    *int       `query:"offset"`
	Order     []es.Order `query:"order"`
	Namespace string     `header:"X-Namespace" required:"true"`
}

func NewFindEntityInteractor[T es.Entity, W any]() usecase.Interactor {
	var entity T
	opts := es.NewEntityOptions(entity)
	entityConfig, err := es.NewEntityConfig(opts)
	if err != nil {
		panic(err)
	}

	whereFactory, err := es.NewWhereFactory[W]()
	if err != nil {
		panic(err)
	}

	items := []T{}

	var in FindInput[W]
	u := usecase.NewIOI(in, items, func(ctx context.Context, input interface{}, output interface{}) error {
		log := xlog.Logger(ctx)

		in := input.(FindInput[W])

		namespace := es.GetNamespace(ctx)
		if len(in.Namespace) > 0 {
			namespace = in.Namespace
		}

		unit, err := es.GetUnit(ctx)
		if err != nil {
			return err
		}

		filter := es.Filter{
			Where:  whereFactory(in.Where),
			Order:  in.Order,
			Limit:  in.Limit,
			Offset: in.Offset,
		}
		if err := unit.Find(ctx, entityConfig.Name, namespace, filter, output); err != nil {
			log.Error("failed to find", zap.String("name", entityConfig.Name), zap.Error(err))
			return fmt.Errorf("failed to find: %w %w", err, status.Unknown)
		}

		return nil
	})

	u.SetTitle(fmt.Sprintf("Find %s", entityConfig.Name))
	u.SetName(fmt.Sprintf("Find %s", entityConfig.Name))
	u.SetExpectedErrors(status.InvalidArgument)
	u.SetExpectedErrors(status.Unknown)
	u.SetExpectedErrors(status.NotFound)
	return u
}
