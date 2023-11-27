package xes

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/google/uuid"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

func NewReplayInteractor(name string) usecase.Interactor {
	type ReplayInput struct {
		Namespace   string    `json:"namespace" required:"true"`
		AggregateId uuid.UUID `json:"aggregate_id" format:"uuid" required:"true"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input ReplayInput, output *Return) error {
		unit, err := es.GetUnit(ctx)
		if err != nil {
			return err
		}

		replay := es.ReplayCommand{
			BaseCommand: es.BaseCommand{
				AggregateId: input.AggregateId,
			},
			BaseNamespaceCommand: es.BaseNamespaceCommand{
				Namespace: input.Namespace,
			},
			AggregateName: name,
		}

		if err := unit.Replay(ctx, &replay); err != nil {
			return err
		}

		output.Id = input.AggregateId
		return nil
	})

	u.SetTitle(fmt.Sprintf("Replay %s", name))
	u.SetName(fmt.Sprintf("Replay.%s", name))
	u.SetExpectedErrors(status.InvalidArgument)
	return u
}
