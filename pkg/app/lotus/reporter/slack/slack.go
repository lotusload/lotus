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

package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
	"github.com/lotusload/lotus/pkg/app/lotus/model"
	"github.com/lotusload/lotus/pkg/app/lotus/reporter"
)

type Message struct {
	Text        string        `json:"text"`
	UserName    string        `json:"username,omitempty"`
	IconURL     string        `json:"icon_url,omitempty"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color      string   `json:"color,omitempty"`
	Title      string   `json:"title,omitempty"`
	TitleLink  string   `json:"title_link,omitempty"`
	Text       string   `json:"text,omitempty"`
	MarkdownIn []string `json:"mrkdwn_in,omitempty"`
}

type builder struct {
}

func NewBuilder() reporter.Builder {
	return &builder{}
}

func (b *builder) Build(r *config.Receiver, opts reporter.BuildOptions) (reporter.Reporter, error) {
	configs, ok := r.Type.(*config.Receiver_Slack)
	if !ok {
		return nil, fmt.Errorf("wrong receiver type for slack: %T", r.Type)
	}
	return &slack{
		hookURL: configs.Slack.HookUrl,
		client:  http.DefaultClient,
		logger:  opts.NamedLogger("slack-reporter"),
	}, nil
}

type slack struct {
	hookURL string
	client  *http.Client
	logger  *zap.Logger
}

func (s *slack) Report(ctx context.Context, result *model.Result) error {
	data, err := result.Render(model.RenderFormatMarkdown)
	if err != nil {
		return err
	}
	att := &Attachment{
		Title: fmt.Sprintf("%s %s", result.TestID, result.Status),
		Text:  fmt.Sprintf("```%s```", string(data)),
		Color: "danger",
		MarkdownIn: []string{
			"text",
		},
	}
	if result.Status == model.TestSucceeded {
		att.Color = "good"
	}
	msg := &Message{
		Attachments: []*Attachment{att},
	}
	if err := s.send(msg); err != nil {
		s.logger.Error("failed to report to slack", zap.Error(err))
		return err
	}
	return nil
}

func (s *slack) send(msg *Message) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, err := s.client.Post(s.hookURL, "application/json", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
