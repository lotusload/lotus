package resource

import (
	"fmt"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
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
