package xes

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
	"github.com/go-apis/utils/xlog"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"go.uber.org/zap"
)

type FindInput interface {
	GetNamespace() string
	GetLimit() *int
	GetOffset() *int
}

type BaseFindInput struct {
	Limit     *int     `query:"limit"`
	Offset    *int     `query:"offset"`
	Order     []string `query:"order"`
	Namespace string   `header:"X-Namespace" required:"true"`
}

func (i *BaseFindInput) GetNamespace() string {
	return i.Namespace
}

func (i *BaseFindInput) GetLimit() *int {
	return i.Limit
}

func (i *BaseFindInput) GetOffset() *int {
	return i.Offset
}

func NewFindEntityInteractor[T es.Entity, W FindInput]() usecase.Interactor {
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
	orderFactory, err := es.NewOrderFactory[W]()
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

		filter := es.Filter{
			Where:  whereFactory(in),
			Order:  orderFactory(in),
			Limit:  in.GetLimit(),
			Offset: in.GetOffset(),
		}

		namespace := in.GetNamespace()
		if len(namespace) == 0 {
			namespace = es.GetNamespace(ctx)
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
