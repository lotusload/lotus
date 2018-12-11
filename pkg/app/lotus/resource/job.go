package resource

import (
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lotusv1beta1 "github.com/nghialv/lotus/pkg/app/lotus/apis/lotus/v1beta1"
	"github.com/nghialv/lotus/pkg/app/lotus/config"
)

type JobType string

const (
	JobPreparer JobType = "preparer"
	JobMonitor          = "monitor"
	JobCleaner          = "cleaner"
)

func newMonitorJob(lotus *lotusv1beta1.Lotus, cfg *config.Config) *batchv1.Job {
	args := []string{
		"monitor",
		fmt.Sprintf("--test-id=%s", lotus.Name),
		fmt.Sprintf("--run-time=%s", lotus.Spec.Worker.RunTime),
		"--config-file=/etc/monitor/config/config.yaml",
		fmt.Sprintf("--collect-summary-datasource=%s", localPrometheusDataSourceName),
	}
	if s := lotus.Spec.CheckIntervalSeconds; s != nil {
		d := time.Duration(*s) * time.Second
		args = append(args, fmt.Sprintf("--check-interval=%s", d.String()))
	}
	if s := lotus.Spec.CheckInitialDelaySeconds; s != nil {
		d := time.Duration(*s) * time.Second
		args = append(args, fmt.Sprintf("--check-initial-delay=%s", d.String()))
	}
	container := corev1.Container{
		Name:  "monitor",
		Image: lotusImage,
		Args:  args,
		Env:   []corev1.EnvVar{},
		VolumeMounts: []corev1.VolumeMount{
			corev1.VolumeMount{
				Name:      "config",
				ReadOnly:  true,
				MountPath: "/etc/monitor/config",
			},
		},
	}
	volumes := []corev1.Volume{
		corev1.Volume{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: jobName(lotus.Name, JobMonitor),
					},
				},
			},
		},
	}
	for _, receiver := range cfg.Receivers {
		gcsReceiver, ok := receiver.Type.(*config.Receiver_Gcs)
		if !ok {
			continue
		}
		if gcsReceiver.Gcs.Credentials != nil {
			volumeName := fmt.Sprintf("gcs-credentials-%s", receiver.Name)
			volumes = append(volumes, corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: gcsReceiver.Gcs.Credentials.Secret,
					},
				},
			})
			path := receiver.CredentialsMountPath()
			container.VolumeMounts = append(container.VolumeMounts,
				corev1.VolumeMount{
					Name:      volumeName,
					MountPath: path,
				},
			)
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  "GOOGLE_APPLICATION_CREDENTIALS",
				Value: fmt.Sprintf("%s%s", path, gcsReceiver.Gcs.Credentials.File),
			})
		}
	}
	return newJob(
		lotus,
		[]corev1.Container{container},
		volumes,
		JobMonitor,
	)
}

func newMonitorConfigMap(lotus *lotusv1beta1.Lotus, config []byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName(lotus.Name, JobMonitor),
			Namespace:       lotus.Namespace,
			OwnerReferences: ownerReferences(lotus),
		},
		BinaryData: map[string][]byte{
			"config.yaml": config,
		},
	}
}

func newJob(lotus *lotusv1beta1.Lotus, containers []corev1.Container, volumes []corev1.Volume, jt JobType) *batchv1.Job {
	var backoffLimit int32
	labels := jobLabels(lotus.Name, jt)
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            jobName(lotus.Name, jt),
			Namespace:       lotus.Namespace,
			OwnerReferences: ownerReferences(lotus),
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers:    containers,
					Volumes:       volumes,
				},
			},
		},
	}
}

func jobName(lotusName string, jt JobType) string {
	return fmt.Sprintf("%s-%s", lotusName, string(jt))
}

func jobLabels(lotusName string, jt JobType) map[string]string {
	return map[string]string{
		"app":      "lotus-job",
		"lotus":    lotusName,
		"job-type": string(jt),
	}
}
