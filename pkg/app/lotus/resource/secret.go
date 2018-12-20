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

package resource

import (
	"fmt"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
)

const (
	timeSeriesStoreConfigFile = "store-config.yaml"
)

type ThanosStore struct {
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

type ThanosGCSConfig struct {
	Bucket string `json:"bucket"`
}

func generateThanosStoreConfig(cfg *config.TimeSeriesStorage) ([]byte, error) {
	switch store := cfg.Type.(type) {
	case *config.TimeSeriesStorage_Gcs:
		return yaml.Marshal(&ThanosStore{
			Type: "GCS",
			Config: &ThanosGCSConfig{
				Bucket: store.Gcs.Bucket,
			},
		})
	default:
		return nil, fmt.Errorf("unsupported store: %v", cfg.Type)
	}
}

func newTimeSeriesStoreConfigSecret(namespace, release string, cfg *config.TimeSeriesStorage, owners []metav1.OwnerReference) (*corev1.Secret, error) {
	data, err := generateThanosStoreConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            timeSeriesStoreConfigSecretName(release),
			Namespace:       namespace,
			OwnerReferences: owners,
		},
		Data: map[string][]byte{
			timeSeriesStoreConfigFile: data,
		},
	}, nil
}

func timeSeriesStoreConfigSecretName(release string) string {
	return fmt.Sprintf("%s-time-series-store-config", release)
}
