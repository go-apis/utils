package xes

import (
	"context"
	"fmt"

	"github.com/go-apis/eventsourcing/es"
	"github.com/google/uuid"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

type Return struct {
	Id uuid.UUID `json:"id" format:"uuid" required:"true"`
}

func NewCommandInteractorExecptedErrors(errors ...error) CommandInteractorOption {
	return func(opts *CommandInteractorOptions) {
		opts.Errors = errors
	}
}

func NewCommandInteractorErrorHandler(handler func(err error) error) CommandInteractorOption {
	return func(opts *CommandInteractorOptions) {
		opts.ErrorHandler = handler
	}
}

type CommandInteractorOption func(opts *CommandInteractorOptions)

type CommandInteractorOptions struct {
	Errors       []error
	ErrorHandler func(err error) error
}

func NewCommandInteractor[T es.Command](options ...CommandInteractorOption) usecase.Interactor {
	opts := &CommandInteractorOptions{
		Errors:       []error{status.InvalidArgument},
		ErrorHandler: func(err error) error { return err },
	}
	for _, option := range options {
		option(opts)
	}

	var cmd T
	commandConfig := es.NewCommandConfig(cmd)

	u := usecase.NewIOI(cmd, new(Return), func(ctx context.Context, input interface{}, output interface{}) error {
		var (
			in  = input.(T)
			out = output.(*Return)
		)

		// do it!
		unit, err := es.GetUnit(ctx)
		if err != nil {
			return opts.ErrorHandler(err)
		}

		if err := unit.Dispatch(ctx, in); err != nil {
			return opts.ErrorHandler(err)
		}

		out.Id = in.GetAggregateId()
		return nil
	})

	u.SetTitle(fmt.Sprintf("Command %s", commandConfig.Name))
	u.SetName(fmt.Sprintf("Command.%s", commandConfig.Name))
	u.SetExpectedErrors(opts.Errors...)
	return u
}
