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

package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
	"github.com/lotusload/lotus/pkg/app/lotus/model"
	"github.com/lotusload/lotus/pkg/app/lotus/reporter"
)

type builder struct {
}

func NewBuilder() reporter.Builder {
	return &builder{}
}

func (b *builder) Build(r *config.Receiver, opts reporter.BuildOptions) (reporter.Reporter, error) {
	return &logger{
		logger: opts.NamedLogger("logger-reporter"),
	}, nil
}

type logger struct {
	logger *zap.Logger
}

func (p *logger) Report(ctx context.Context, result *model.Result) error {
	out, err := result.Render(model.RenderFormatText)
	if err != nil {
		return err
	}
	fmt.Println("=========== TEST RESULT ==========")
	fmt.Println(string(out))
	fmt.Println("===========     END     ==========")
	return nil
}
