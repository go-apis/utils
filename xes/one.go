package xes

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/utils/xlog"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"go.uber.org/zap"
)

type OneInput interface {
	GetNamespace() string
}

type BaseOneInput struct {
	Namespace string `header:"X-Namespace" required:"true"`
}

func (i *BaseOneInput) GetNamespace() string {
	return i.Namespace
}

func NewOneEntityInteractor[T es.Entity, W OneInput]() usecase.Interactor {
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

	var in W
	u := usecase.NewIOI(in, items, func(ctx context.Context, input interface{}, output interface{}) error {
		in := input.(W)

		unit, err := es.GetUnit(ctx)
		if err != nil {
			return err
		}

		log := xlog.Logger(ctx)

		one := 1
		filter := es.Filter{
			Where: whereFactory(in),
			Limit: &one,
		}

		namespace := in.GetNamespace()
		if len(namespace) == 0 {
			namespace = es.GetNamespace(ctx)
		}

		errOne := unit.One(ctx, entityConfig.Name, namespace, filter, output)
		switch {
		case errOne == nil:
			return nil
		case errOne == sql.ErrNoRows:
			return status.NotFound
		default:
			log.Error("failed to find", zap.String("name", entityConfig.Name), zap.Error(err))
			return fmt.Errorf("failed to find: %w %w", err, status.Unknown)
		}
	})

	u.SetTitle(fmt.Sprintf("One %s", entityConfig.Name))
	u.SetName(fmt.Sprintf("One %s", entityConfig.Name))
	u.SetExpectedErrors(status.InvalidArgument)
	u.SetExpectedErrors(status.Unknown)
	u.SetExpectedErrors(status.NotFound)
	return u
}
