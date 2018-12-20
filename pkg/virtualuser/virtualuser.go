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

package virtualuser

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

type (
	VirtualUser interface {
		Run(ctx context.Context) error
	}

	Factory func() (VirtualUser, error)
)

const (
	StatusStarted   = "started"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

var (
	UserCount = stats.Int64("virtual_user/count", "Number of virtual users", stats.UnitDimensionless)

	KeyStatus, _ = tag.NewKey("virtual_user_status")

	UserCountView = &view.View{
		Name:        "virtual_user/count",
		Measure:     UserCount,
		TagKeys:     []tag.Key{KeyStatus},
		Description: "Count of virtual users by status",
		Aggregation: view.Count(),
	}
)

type Group struct {
	numUsers  int
	hatchRate int
	factory   Factory
	doneCh    chan struct{}
}

func NewGroup(numUsers, hatchRate int, factory Factory) *Group {
	return &Group{
		numUsers:  numUsers,
		hatchRate: hatchRate,
		factory:   factory,
		doneCh:    make(chan struct{}),
	}
}

func (g *Group) Run(ctx context.Context) error {
	startedCtx, err := tag.New(ctx, tag.Insert(KeyStatus, StatusStarted))
	if err != nil {
		return err
	}
	succeededCtx, err := tag.New(ctx, tag.Insert(KeyStatus, StatusSucceeded))
	if err != nil {
		return err
	}
	failedCtx, err := tag.New(ctx, tag.Insert(KeyStatus, StatusFailed))
	if err != nil {
		return err
	}
	var total int = 0
	var wg sync.WaitGroup
	for total < g.numUsers {
		for i := 0; i < g.hatchRate && total < g.numUsers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				stats.Record(startedCtx, UserCount.M(1))
				var err error
				defer func() {
					if r := recover(); r != nil {
						err = errors.New("virtualuser.group: panic")
					}
					if err != nil {
						stats.Record(succeededCtx, UserCount.M(1))
					} else {
						stats.Record(failedCtx, UserCount.M(1))
					}
				}()
				var vu VirtualUser
				vu, err = g.factory()
				if err != nil {
					return
				}
				err = vu.Run(ctx)
			}()
			total++
		}
		if total >= g.numUsers {
			break
		}
		select {
		case <-ctx.Done():
			break
		case <-time.After(time.Second):
		}
	}
	wg.Wait()
	close(g.doneCh)
	return nil
}

func (g *Group) Stop(timeout time.Duration) error {
	select {
	case <-g.doneCh:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timed out")
	}
}
