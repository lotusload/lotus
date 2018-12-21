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

package cli

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRunWithContext(t *testing.T) {
	calls := 0
	runner := func(ctx context.Context, logger *zap.Logger) error {
		calls++
		timeout := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			return nil
		case <-timeout.C:
			return fmt.Errorf("timed out")
		}
	}
	ch := make(chan os.Signal, 1)
	ch <- syscall.SIGINT
	err := runWithContext(&cobra.Command{}, ch, runner)
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestNewLogger(t *testing.T) {
	logger, err := newLogger(&cobra.Command{})
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}
