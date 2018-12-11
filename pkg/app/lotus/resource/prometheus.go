package resource

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	lotusv1beta1 "github.com/nghialv/lotus/pkg/app/lotus/apis/lotus/v1beta1"
	"github.com/nghialv/lotus/pkg/app/lotus/config"
)

const (
	prometheusConfigDirectory     = "/etc/prometheus"
	prometheusConfigFile          = "prometheus-config.yaml"
	prometheusRuleFile            = "prometheus-rule.yaml"
	prometheusPort                = 9090
	localPrometheusDataSourceName = "_LocalPrometheus"
	prometheusBlockDuration       = "1m"
)

func newPrometheusPod(lotus *lotusv1beta1.Lotus, serviceAccount, release string, cfg *config.Config) (*corev1.Pod, error) {
	volumes := []corev1.Volume{
		corev1.Volume{
			Name: "db",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		corev1.Volume{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: prometheusName(lotus.Name),
					},
				},
			},
		},
	}
	prometheusContainer := corev1.Container{
		Name:  "prometheus",
		Image: prometheusImage,
		Args: []string{
			fmt.Sprintf("--config.file=%s/%s", prometheusConfigDirectory, prometheusConfigFile),
			"--storage.tsdb.path=/var/prometheus",
			fmt.Sprintf("--storage.tsdb.min-block-duration=%s", prometheusBlockDuration),
			fmt.Sprintf("--storage.tsdb.max-block-duration=%s", prometheusBlockDuration),
			"--storage.tsdb.retention=6h",
			"--web.enable-lifecycle",
		},
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				Name:          "prom-http",
				ContainerPort: prometheusPort,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "config",
				MountPath: prometheusConfigDirectory,
			},
			corev1.VolumeMount{
				Name:      "db",
				MountPath: "/var/prometheus",
			},
		},
	}
	thanosContainer := corev1.Container{
		Name:  "thanos-sidecar",
		Image: thanosImage,
		Args: []string{
			"sidecar",
			"--tsdb.path=/var/prometheus",
			fmt.Sprintf("--prometheus.url=http://127.0.0.1:%d", prometheusPort),
			"--cluster.disable",
		},
		Env:   []corev1.EnvVar{},
		Ports: thanosPorts(),
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "config",
				MountPath: prometheusConfigDirectory,
			},
			corev1.VolumeMount{
				Name:      "db",
				MountPath: "/var/prometheus",
			},
		},
	}
	if cfg.TimeSeriesStorage != nil {
		setTimeSeriesStoreConfig(&thanosContainer, &volumes, release)
		if gcs, ok := cfg.TimeSeriesStorage.Type.(*config.TimeSeriesStorage_Gcs); ok {
			if gcs.Gcs.Credentials != nil {
				setGCSCredentials(&thanosContainer, &volumes, gcs.Gcs.Credentials)
			}
		}
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            prometheusName(lotus.Name),
			Namespace:       lotus.Namespace,
			OwnerReferences: ownerReferences(lotus),
			Labels:          prometheusPodLabels(lotus.Name, release),
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				prometheusContainer,
				thanosContainer,
			},
			Volumes: volumes,
		},
	}
	if serviceAccount != "" {
		pod.Spec.ServiceAccountName = serviceAccount
	}
	return pod, nil
}

func newPrometheusService(lotus *lotusv1beta1.Lotus) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            prometheusName(lotus.Name),
			Namespace:       lotus.Namespace,
			OwnerReferences: ownerReferences(lotus),
		},
		Spec: corev1.ServiceSpec{
			Selector: prometheusServiceLabels(lotus.Name),
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "metrics",
					TargetPort: intstr.FromInt(prometheusPort),
					Port:       int32(prometheusPort),
				},
			},
		},
	}
}

func newPrometheusConfigMap(lotus *lotusv1beta1.Lotus, target string, globalChecks []lotusv1beta1.LotusCheck) (*corev1.ConfigMap, error) {
	config, err := renderTemplate(
		&prometheusConfigParams{
			Name:        prometheusName(lotus.Name),
			Namespace:   lotus.Namespace,
			ServiceName: target,
			RuleFiles: []string{
				fmt.Sprintf("%s/%s", prometheusConfigDirectory, prometheusRuleFile),
			},
		},
		prometheusConfigTemplate,
	)
	if err != nil {
		return nil, err
	}
	globalChecks = append(globalChecks, lotus.Spec.Checks...)
	rule, err := renderTemplate(
		&prometheusRuleParams{
			Alerts: globalChecks,
		},
		prometheusRuleTemplate,
	)
	if err != nil {
		return nil, err
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            prometheusName(lotus.Name),
			Namespace:       lotus.Namespace,
			OwnerReferences: ownerReferences(lotus),
		},
		BinaryData: map[string][]byte{
			prometheusConfigFile: config,
			prometheusRuleFile:   rule,
		},
	}, nil
}

func prometheusName(lotusName string) string {
	return fmt.Sprintf("%s-prometheus", lotusName)
}

func prometheusServiceLabels(lotusName string) map[string]string {
	return map[string]string{
		"app":   "lotus-prometheus",
		"lotus": lotusName,
	}
}

func prometheusPodLabels(lotusName, release string) map[string]string {
	return map[string]string{
		"app":           "lotus-prometheus",
		"lotus":         lotusName,
		thanosPeerLabel: release,
	}
}

func clientPrometheusDataSource(lotus *lotusv1beta1.Lotus) *config.DataSource {
	address := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		prometheusName(lotus.Name),
		lotus.Namespace,
		prometheusPort,
	)
	return &config.DataSource{
		Name: localPrometheusDataSourceName,
		Type: &config.DataSource_Prometheus{
			Prometheus: &config.PrometheusConfigs{
				Address: address,
			},
		},
	}
}
