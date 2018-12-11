package reporter

import (
	"context"
	"errors"
	"testing"

	"github.com/nghialv/lotus/pkg/app/lotus/model"
	"github.com/stretchr/testify/assert"
)

type reporterFunc func(ctx context.Context, result *model.Result) error

func (f reporterFunc) Report(ctx context.Context, result *model.Result) error {
	return f(ctx, result)
}

func TestMultiReporter(t *testing.T) {
	var calls int
	successReporter := reporterFunc(func(ctx context.Context, result *model.Result) error {
		calls++
		return nil
	})
	failureReporter := reporterFunc(func(ctx context.Context, result *model.Result) error {
		calls++
		return errors.New("failed")
	})
	testcases := []struct {
		reporter Reporter
		result   *model.Result
		hasError bool
		calls    int
	}{
		{
			reporter: MultiReporter(successReporter),
			hasError: false,
			calls:    1,
		},
		{
			reporter: MultiReporter(successReporter, successReporter),
			hasError: false,
			calls:    2,
		},
		{
			reporter: MultiReporter(successReporter, failureReporter),
			hasError: true,
			calls:    2,
		},
		{
			reporter: MultiReporter(successReporter, MultiReporter(successReporter, successReporter)),
			hasError: false,
			calls:    3,
		},
		{
			reporter: MultiReporter(successReporter, MultiReporter(successReporter, failureReporter)),
			hasError: true,
			calls:    3,
		},
	}
	for _, tc := range testcases {
		calls = 0
		err := tc.reporter.Report(context.TODO(), tc.result)
		assert.Equal(t, tc.hasError, err != nil)
		assert.Equal(t, tc.calls, calls)
	}
}
