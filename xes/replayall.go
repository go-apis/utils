package xes

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"
	"github.com/contextcloud/eventstore/es/filters"
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
		events, err := unit.FindEvents(nctx, filters.Filter{
			Where: []filters.WhereClause{
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

		replayCmds := make([]*es.ReplayCommand, len(events))
		for i, evt := range events {
			replayCmds[i] = &es.ReplayCommand{
				BaseCommand: es.BaseCommand{
					AggregateId: evt.AggregateId,
				},
				BaseNamespaceCommand: es.BaseNamespaceCommand{
					Namespace: evt.Namespace,
				},
				AggregateName: evt.AggregateType,
			}
		}

		if err := unit.Replay(ctx, replayCmds...); err != nil {
			return err
		}

		output.TotalCommands = len(replayCmds)
		return nil
	})

	u.SetTitle(fmt.Sprintf("Replay All %s", name))
	u.SetName(fmt.Sprintf("ReplayAll.%s", name))
	u.SetExpectedErrors(status.InvalidArgument)
	return u
}
