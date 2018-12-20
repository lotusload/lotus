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

package gcs

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"google.golang.org/api/option"

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
	configs, ok := r.Type.(*config.Receiver_Gcs)
	if !ok {
		return nil, fmt.Errorf("wrong receiver type for gcs: %T", r.Type)
	}
	return &gcs{
		bucket:          configs.Gcs.Bucket,
		credentialsFile: r.CredentialsFile(configs.Gcs.Credentials.File),
		logger:          opts.NamedLogger("gcs-reporter"),
	}, nil
}

type gcs struct {
	bucket          string
	credentialsFile string
	logger          *zap.Logger
}

func (g *gcs) Report(ctx context.Context, result *model.Result) (lastErr error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(g.credentialsFile))
	if err != nil {
		g.logger.Error("failed to create gcs storage client", zap.Error(err))
		lastErr = err
		return
	}
	cases := []struct {
		format      model.RenderFormat
		extension   string
		contentType string
	}{
		{
			format:      model.RenderFormatText,
			extension:   "txt",
			contentType: "text/plain",
		},
		{
			format:      model.RenderFormatJson,
			extension:   "json",
			contentType: "application/json",
		},
	}
	for _, c := range cases {
		data, err := result.Render(c.format)
		if err != nil {
			g.logger.Error("failed to render result", zap.Error(err))
			lastErr = err
			continue
		}
		filename := fmt.Sprintf("%s/%s.%s", result.TestID, result.TestID, c.extension)
		g.logger.Info("writing test result to gcs storage",
			zap.String("testID", result.TestID),
			zap.String("filename", filename),
			zap.String("bucket", g.bucket),
		)
		wc := client.Bucket(g.bucket).Object(filename).NewWriter(ctx)
		wc.ContentType = c.contentType
		wc.ACL = []storage.ACLRule{{
			Entity: storage.AllUsers,
			Role:   storage.RoleReader,
		}}
		if _, err := wc.Write(data); err != nil {
			g.logger.Error("failed to write result", zap.Error(err))
			lastErr = err
			continue
		}
		if err := wc.Close(); err != nil {
			g.logger.Error("failed to close writer", zap.Error(err))
			lastErr = err
			continue
		}
	}
	return
}
