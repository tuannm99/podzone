package messaging

import (
	"context"
	"fmt"
)

type Saga struct {
	steps []SagaStep
}

func NewSaga(steps ...SagaStep) *Saga {
	return &Saga{steps: append([]SagaStep(nil), steps...)}
}

func (s *Saga) Run(ctx context.Context, msg Envelope) error {
	completed := make([]SagaStep, 0, len(s.steps))
	for _, step := range s.steps {
		if err := step.Execute(ctx, msg); err != nil {
			for i := len(completed) - 1; i >= 0; i-- {
				_ = completed[i].Compensate(ctx, msg)
			}
			return fmt.Errorf("saga step %s failed: %w", step.Name(), err)
		}
		completed = append(completed, step)
	}
	return nil
}
