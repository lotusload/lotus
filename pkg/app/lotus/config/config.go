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

package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"

	lotusv1beta1 "github.com/lotusload/lotus/pkg/app/lotus/apis/lotus/v1beta1"
)

func (c *Config) AddChecks(checks ...lotusv1beta1.LotusCheck) {
	for i := range checks {
		c.Checks = append(c.Checks, &Check{
			Name:       checks[i].Name,
			Expr:       checks[i].Expr,
			For:        checks[i].For,
			DataSource: checks[i].DataSource,
		})
	}
}

func (c *Config) LotusChecks() []lotusv1beta1.LotusCheck {
	checks := make([]lotusv1beta1.LotusCheck, 0, len(c.Checks))
	for _, check := range c.Checks {
		checks = append(checks, lotusv1beta1.LotusCheck{
			Name:       check.Name,
			Expr:       check.Expr,
			For:        check.For,
			DataSource: check.DataSource,
		})
	}
	return checks
}

func (ds *DataSource) DataSourceType() DataSource_Type {
	switch ds.Type.(type) {
	case *DataSource_Prometheus:
		return DataSource_PROMETHEUS
	default:
		return DataSource_UNKNOWN
	}
}

func (r *Receiver) ReceiverType() Receiver_Type {
	switch r.Type.(type) {
	case *Receiver_Logger:
		return Receiver_LOGGER
	case *Receiver_Gcs:
		return Receiver_GCS
	case *Receiver_Slack:
		return Receiver_SLACK
	default:
		return Receiver_UNKNOWN
	}
}

func (r *Receiver) CredentialsMountPath() string {
	return fmt.Sprintf("/etc/creds/%s/", r.Name)
}

func (r *Receiver) CredentialsFile(filename string) string {
	return fmt.Sprintf("%s%s", r.CredentialsMountPath(), filename)
}

func FromFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return UnmarshalFromYaml(data)
}

func UnmarshalFromYaml(data []byte) (*Config, error) {
	json, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	if err = jsonpb.UnmarshalString(string(json), config); err != nil {
		return nil, err
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) MarshalToYaml() ([]byte, error) {
	marshaler := &jsonpb.Marshaler{}
	json, err := marshaler.MarshalToString(c)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(json))
}
