package containertest

import (
	"context"

	"agentrun/internal/container"
)

type Fake struct {
	Results  []container.Result
	StartErr error
	Specs    []container.Spec
}

func (f *Fake) Run(ctx context.Context, spec container.Spec) (container.Result, error) {
	if err := spec.Validate(); err != nil {
		return container.Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return container.Result{}, err
	}
	f.Specs = append(f.Specs, spec)
	if f.StartErr != nil {
		return container.Result{}, f.StartErr
	}
	if len(f.Results) == 0 {
		return container.Result{}, nil
	}
	result := f.Results[0]
	f.Results = f.Results[1:]
	return result, nil
}
