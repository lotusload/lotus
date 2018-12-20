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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lotusload/lotus/pkg/app/lotus/config"
)

type StaticResourceFactory interface {
	ThanosStoreName() string
	ThanosQueryName() string
	ThanosPeerName() string
	TimeSeriesStoreConfigSecretName() string

	NewThanosStoreStatefulSet() (*appsv1.StatefulSet, error)
	NewThanosQueryDeployment() (*appsv1.Deployment, error)
	NewThanosQueryService() (*corev1.Service, error)
	NewThanosPeerService() (*corev1.Service, error)
	NewTimeSeriesStoreConfigSecret() (*corev1.Secret, error)
}

type staticResourceFactory struct {
	namespace       string
	release         string
	configFile      string
	ownerReferences []metav1.OwnerReference
}

func NewStaticResourceFactory(namespace, release, configFile string, owners []metav1.OwnerReference) StaticResourceFactory {
	return &staticResourceFactory{
		namespace:       namespace,
		release:         release,
		configFile:      configFile,
		ownerReferences: owners,
	}
}

func (f *staticResourceFactory) ThanosStoreName() string {
	return thanosStoreName(f.release)
}

func (f *staticResourceFactory) ThanosQueryName() string {
	return thanosQueryName(f.release)
}

func (f *staticResourceFactory) ThanosPeerName() string {
	return thanosPeerName(f.release)
}

func (f *staticResourceFactory) TimeSeriesStoreConfigSecretName() string {
	return timeSeriesStoreConfigSecretName(f.release)
}

func (f *staticResourceFactory) NewThanosStoreStatefulSet() (*appsv1.StatefulSet, error) {
	cfg, err := config.FromFile(f.configFile)
	if err != nil {
		return nil, err
	}
	return newThanosStoreStatefulSet(f.namespace, f.release, cfg.TimeSeriesStorage, f.ownerReferences)
}

func (f *staticResourceFactory) NewThanosQueryDeployment() (*appsv1.Deployment, error) {
	return newThanosQueryDeployment(f.namespace, f.release, f.ownerReferences), nil
}

func (f *staticResourceFactory) NewThanosQueryService() (*corev1.Service, error) {
	return newThanosQueryService(f.namespace, f.release, f.ownerReferences), nil
}

func (f *staticResourceFactory) NewThanosPeerService() (*corev1.Service, error) {
	return newThanosPeerService(f.namespace, f.release, f.ownerReferences), nil
}

func (f *staticResourceFactory) NewTimeSeriesStoreConfigSecret() (*corev1.Secret, error) {
	cfg, err := config.FromFile(f.configFile)
	if err != nil {
		return nil, err
	}
	return newTimeSeriesStoreConfigSecret(f.namespace, f.release, cfg.TimeSeriesStorage, f.ownerReferences)
}
