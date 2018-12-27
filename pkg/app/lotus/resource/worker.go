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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	lotusv1beta1 "github.com/lotusload/lotus/pkg/app/lotus/apis/lotus/v1beta1"
)

func newWorkerDeployment(lotus *lotusv1beta1.Lotus) *appsv1.Deployment {
	labels := workerLabels(lotus.Name)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            workerName(lotus.Name),
			Namespace:       lotus.Namespace,
			OwnerReferences: ownerReferences(lotus),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: lotus.Spec.Worker.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers:    lotus.Spec.Worker.Containers,
					Volumes:       lotus.Spec.Worker.Volumes,
				},
			},
		},
	}
}

func newWorkerService(lotus *lotusv1beta1.Lotus) *corev1.Service {
	labels := workerLabels(lotus.Name)
	metricsPort := *lotus.Spec.Worker.MetricsPort
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            workerName(lotus.Name),
			Namespace:       lotus.Namespace,
			OwnerReferences: ownerReferences(lotus),
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "metrics",
					TargetPort: intstr.FromInt(int(metricsPort)),
					Port:       metricsPort,
				},
			},
		},
	}

}

func workerName(lotusName string) string {
	return fmt.Sprintf("%s-worker", lotusName)
}

func workerLabels(lotusName string) map[string]string {
	return map[string]string{
		"app":   "lotus-worker",
		"lotus": lotusName,
	}
}
