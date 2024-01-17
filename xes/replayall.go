package xes

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

func NewReplayAllInteractor(name string) usecase.Interactor {
	type ReplayAllInput struct {
		Namespace string `json:"namespace" required:"true"`
	}
	type ReplayAllOutput struct {
		TotalCommands int `json:"total_commands"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input ReplayAllInput, output *ReplayAllOutput) error {
		unit, err := es.GetUnit(ctx)
		if err != nil {
			return err
		}

		nctx := es.SetNamespace(ctx, input.Namespace)
		events, err := unit.FindEvents(nctx, es.Filter{
			Where: []es.WhereClause{
				{
					Column: "aggregate_type",
					Op:     "eq",
					Args:   name,
				},
				{
					Column: "version",
					Op:     "eq",
					Args:   1,
				},
			},
		})
		if err != nil {
			return err
		}

		cmds := make([]es.Command, len(events))
		for i, evt := range events {
			cmds[i] = es.NewReplayCommand(evt.Namespace, evt.AggregateId, evt.AggregateType)
		}

		if err := unit.Dispatch(ctx, cmds...); err != nil {
			return err
		}

		output.TotalCommands = len(cmds)
		return nil
	})

	u.SetTitle(fmt.Sprintf("Replay All %s", name))
	u.SetName(fmt.Sprintf("ReplayAll.%s", name))
	u.SetExpectedErrors(status.InvalidArgument)
	return u
}
