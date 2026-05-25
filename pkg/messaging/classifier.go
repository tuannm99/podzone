package messaging

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type FailureAction string

const (
	FailureActionReturn     FailureAction = "return"
	FailureActionRetry      FailureAction = "retry"
	FailureActionDeadLetter FailureAction = "dead_letter"
	FailureActionDrop       FailureAction = "drop"
)

type FailureClassification struct {
	Action FailureAction
	Reason string
	Err    error
}

type ErrorClassifier interface {
	Classify(ctx context.Context, msg Envelope, err error) FailureClassification
}

type ErrorClassifierFunc func(ctx context.Context, msg Envelope, err error) FailureClassification

func (f ErrorClassifierFunc) Classify(ctx context.Context, msg Envelope, err error) FailureClassification {
	return f(ctx, msg, err)
}

type classifiedError struct {
	action FailureAction
	reason string
	err    error
}

func (e *classifiedError) Error() string {
	if e.err == nil {
		return string(e.action)
	}
	if strings.TrimSpace(e.reason) == "" {
		return e.err.Error()
	}
	return fmt.Sprintf("%s: %v", e.reason, e.err)
}

func (e *classifiedError) Unwrap() error {
	return e.err
}

func RetryableError(err error, reason string) error {
	return &classifiedError{action: FailureActionRetry, reason: reason, err: err}
}

func DeadLetterError(err error, reason string) error {
	return &classifiedError{action: FailureActionDeadLetter, reason: reason, err: err}
}

func DropError(err error, reason string) error {
	return &classifiedError{action: FailureActionDrop, reason: reason, err: err}
}

func ClassifyError(err error) FailureClassification {
	if err == nil {
		return FailureClassification{}
	}
	var classified *classifiedError
	if errors.As(err, &classified) {
		return FailureClassification{
			Action: classified.action,
			Reason: strings.TrimSpace(classified.reason),
			Err:    err,
		}
	}
	return FailureClassification{
		Action: FailureActionReturn,
		Err:    err,
	}
}

func DefaultErrorClassifier() ErrorClassifier {
	return ErrorClassifierFunc(func(_ context.Context, _ Envelope, err error) FailureClassification {
		classification := ClassifyError(err)
		if classification.Action == "" {
			classification.Action = FailureActionReturn
		}
		return classification
	})
}
