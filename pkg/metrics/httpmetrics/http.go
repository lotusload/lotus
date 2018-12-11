// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpmetrics

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

type Transport struct {
	Base           http.RoundTripper
	UsePathAsRoute bool
}

func ContextWithRoute(ctx context.Context, route string) (context.Context, error) {
	return tag.New(ctx, tag.Upsert(KeyClientRoute, route))
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	mutators := []tag.Mutator{
		tag.Upsert(KeyClientHost, req.URL.Host),
		tag.Upsert(KeyClientMethod, req.Method),
	}
	if t.UsePathAsRoute {
		mutators = append(mutators, tag.Upsert(KeyClientRoute, req.URL.Path))
	} else {
		mutators = append(mutators, tag.Insert(KeyClientRoute, "no_route"))
	}
	ctx, _ := tag.New(req.Context(), mutators...)
	req = req.WithContext(ctx)
	track := &tracker{
		start: time.Now(),
		ctx:   ctx,
	}
	track.reqSize = req.ContentLength
	if req.Body == nil && req.ContentLength == -1 {
		track.reqSize = 0
	}

	resp, err := t.base().RoundTrip(req)
	if err != nil {
		track.statusCode = http.StatusInternalServerError
		track.end()
	} else {
		track.statusCode = resp.StatusCode
		track.respContentLength = resp.ContentLength
		if resp.Body == nil {
			track.end()
		} else {
			track.body = resp.Body
			resp.Body = track
		}
	}
	return resp, err
}

func (t *Transport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := t.base().(canceler); ok {
		cr.CancelRequest(req)
	}
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

type tracker struct {
	ctx               context.Context
	respSize          int64
	respContentLength int64
	reqSize           int64
	start             time.Time
	body              io.ReadCloser
	statusCode        int
	endOnce           sync.Once
}

var _ io.ReadCloser = (*tracker)(nil)

func (t *tracker) end() {
	t.endOnce.Do(func() {
		latencyMs := float64(time.Since(t.start)) / float64(time.Millisecond)
		respSize := t.respSize
		if t.respSize == 0 && t.respContentLength > 0 {
			respSize = t.respContentLength
		}
		m := []stats.Measurement{
			ClientSentBytes.M(t.reqSize),
			ClientReceivedBytes.M(respSize),
			ClientRoundtripLatency.M(latencyMs),
		}
		stats.RecordWithTags(t.ctx, []tag.Mutator{
			tag.Upsert(KeyClientStatus, strconv.Itoa(t.statusCode)),
		}, m...)
	})
}

func (t *tracker) Read(b []byte) (int, error) {
	n, err := t.body.Read(b)
	t.respSize += int64(n)
	switch err {
	case nil:
		return n, nil
	case io.EOF:
		t.end()
	}
	return n, err
}

func (t *tracker) Close() error {
	t.end()
	return t.body.Close()
}
