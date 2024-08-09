package xes

import (
	"context"
	"fmt"
	"math"

	"github.com/contextcloud/eventstore/es"
	"github.com/go-apis/utils/xlog"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"go.uber.org/zap"
)

type Pagination[T any] struct {
	PageSize   int   `json:"page_size" required:"true"`
	Page       int   `json:"page" required:"true"`
	TotalItems int64 `json:"total_items" required:"true"`
	TotalPages int   `json:"total_pages" required:"true"`
	Items      []T   `json:"items" required:"true"`
}

type PagingInput interface {
	GetNamespace() string
	GetPage() int
	GetPageSize() int
}

type BasePagingInput struct {
	Namespace string   `header:"X-Namespace" required:"true"`
	Page      int      `query:"page" default:"1"`
	PageSize  int      `query:"page_size" default:"10"`
	Order     []string `query:"order"`
}

func (i *BasePagingInput) GetNamespace() string {
	return i.Namespace
}

func (i *BasePagingInput) GetPage() int {
	return i.Page
}

func (i *BasePagingInput) GetPageSize() int {
	return i.PageSize
}

func NewPagingEntityInteractor[T es.Entity, W PagingInput]() usecase.Interactor {
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

	items := &Pagination[T]{}
	var in W
	u := usecase.NewIOI(in, items, func(ctx context.Context, input interface{}, output interface{}) error {
		in := input.(W)
		out := output.(*Pagination[T])

		page := in.GetPage()
		pageSize := in.GetPageSize()

		offset := (page - 1) * pageSize
		filter := es.Filter{
			Where:  whereFactory(in),
			Order:  orderFactory(in),
			Limit:  &pageSize,
			Offset: &offset,
		}

		namespace := in.GetNamespace()
		if len(namespace) == 0 {
			namespace = es.GetNamespace(ctx)
		}

		unit, err := es.GetUnit(ctx)
		if err != nil {
			return err
		}

		log := xlog.Logger(ctx)

		totalItems, err := unit.Count(ctx, entityConfig.Name, namespace, filter)
		if err != nil {
			log.Error("failed to count", zap.String("name", entityConfig.Name), zap.Error(err))
			return fmt.Errorf("failed to count: %w %w", err, status.Unknown)
		}
		totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

		var items []T
		if err := unit.Find(ctx, entityConfig.Name, namespace, filter, &items); err != nil {
			log.Error("failed to find", zap.String("name", entityConfig.Name), zap.Error(err))
			return fmt.Errorf("failed to find: %w %w", err, status.Unknown)
		}

		out.PageSize = pageSize
		out.Page = page
		out.TotalItems = int64(totalItems)
		out.TotalPages = totalPages
		out.Items = items
		return nil
	})

	u.SetTitle(fmt.Sprintf("Paging %s", entityConfig.Name))
	u.SetName(fmt.Sprintf("Paging %s", entityConfig.Name))
	u.SetExpectedErrors(status.InvalidArgument)
	u.SetExpectedErrors(status.Unknown)
	u.SetExpectedErrors(status.NotFound)
	return u
}
