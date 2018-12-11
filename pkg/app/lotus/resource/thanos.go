package resource

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/nghialv/lotus/pkg/app/lotus/config"
)

const (
	thanosPeerLabel = "lotus-thanos-peer"
)

func newThanosStoreStatefulSet(namespace, release string, cfg *config.TimeSeriesStorage, owners []metav1.OwnerReference) (*appsv1.StatefulSet, error) {
	volumes := []corev1.Volume{
		corev1.Volume{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	container := corev1.Container{
		Name:  "thanos-store",
		Image: thanosImage,
		Args: []string{
			"store",
			"--data-dir=/var/thanos/store",
			"--cluster.disable",
		},
		Env:   []corev1.EnvVar{},
		Ports: thanosPorts(),
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "data",
				MountPath: "/var/thanos/store",
			},
		},
	}
	if cfg != nil {
		setTimeSeriesStoreConfig(&container, &volumes, release)
		if gcs, ok := cfg.Type.(*config.TimeSeriesStorage_Gcs); ok {
			if gcs.Gcs.Credentials != nil {
				setGCSCredentials(&container, &volumes, gcs.Gcs.Credentials)
			}
		}
	}

	labels := thanosStoreLabels(release)
	replicas := int32(1)

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            thanosStoreName(release),
			Namespace:       namespace,
			OwnerReferences: owners,
			Labels:          labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
					Volumes:    volumes,
				},
			},
		},
	}
	return statefulset, nil
}

func newThanosPeerService(namespace, release string, owners []metav1.OwnerReference) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            thanosPeerName(release),
			Namespace:       namespace,
			OwnerReferences: owners,
		},
		Spec: corev1.ServiceSpec{
			Selector:  thanosPeerLabels(release),
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "grpc",
					TargetPort: intstr.FromString("grpc"),
					Port:       int32(10901),
				},
			},
		},
	}
}

func newThanosQueryDeployment(namespace, release string, owners []metav1.OwnerReference) *appsv1.Deployment {
	replicas := int32(1)
	labels := thanosQueryLabels(release)
	containers := []corev1.Container{
		corev1.Container{
			Name:  "thanos-query",
			Image: thanosImage,
			Args: []string{
				"query",
				"--query.replica-label=replica",
				"--cluster.disable",
				fmt.Sprintf("--store=dns+%s.%s.svc.cluster.local:10901", thanosPeerName(release), namespace),
			},
			Ports: thanosPorts(),
		},
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            thanosQueryName(release),
			Namespace:       namespace,
			OwnerReferences: owners,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers:    containers,
				},
			},
		},
	}
}

func newThanosQueryService(namespace, release string, owners []metav1.OwnerReference) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            thanosQueryName(release),
			Namespace:       namespace,
			OwnerReferences: owners,
		},
		Spec: corev1.ServiceSpec{
			Selector: thanosQueryLabels(release),
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "http",
					TargetPort: intstr.FromString("http"),
					Port:       int32(9090),
				},
			},
		},
	}
}

func thanosStoreName(release string) string {
	return fmt.Sprintf("%s-thanos-store", release)
}

func thanosPeerName(release string) string {
	return fmt.Sprintf("%s-thanos-peers", release)
}

func thanosQueryName(release string) string {
	return fmt.Sprintf("%s-thanos-query", release)
}

func thanosStoreLabels(release string) map[string]string {
	return map[string]string{
		"app":           thanosStoreName(release),
		thanosPeerLabel: release,
	}
}

func thanosPeerLabels(release string) map[string]string {
	return map[string]string{
		thanosPeerLabel: release,
	}
}

func thanosQueryLabels(release string) map[string]string {
	return map[string]string{
		"app": thanosQueryName(release),
	}
}

func thanosPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          "http",
			ContainerPort: 10902,
		},
		corev1.ContainerPort{
			Name:          "grpc",
			ContainerPort: 10901,
		},
	}
}

func setGCSCredentials(container *corev1.Container, volumes *[]corev1.Volume, credentials *config.SecretFileSelector) {
	container.Env = append(container.Env, corev1.EnvVar{
		Name:  "GOOGLE_APPLICATION_CREDENTIALS",
		Value: fmt.Sprintf("/creds/gcs/%s", credentials.File),
	})
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      "gcs-credentials",
		MountPath: "/creds/gcs/",
	})
	*volumes = append(*volumes, corev1.Volume{
		Name: "gcs-credentials",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: credentials.Secret,
			},
		},
	})
}

func setTimeSeriesStoreConfig(container *corev1.Container, volumes *[]corev1.Volume, release string) {
	container.Args = append(container.Args,
		fmt.Sprintf("--objstore.config-file=/creds/objstore/%s", timeSeriesStoreConfigFile),
	)
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      "time-series-store-config",
		MountPath: "/creds/objstore/",
	})
	*volumes = append(*volumes, corev1.Volume{
		Name: "time-series-store-config",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: timeSeriesStoreConfigSecretName(release),
			},
		},
	})
}
