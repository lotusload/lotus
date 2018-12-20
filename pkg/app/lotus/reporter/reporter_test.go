// Copyright (c) 2018 Lotus Load
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package reporter

import (
	"context"
	"errors"
	"testing"

	"github.com/lotusload/lotus/pkg/app/lotus/model"
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
