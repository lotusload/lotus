package kubeclient

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"
)

type KubeClient interface {
	EnsureDeployment(name, namespace string, factory func() (*appsv1.Deployment, error)) (*appsv1.Deployment, error)
	EnsurePod(name, namespace string, factory func() (*corev1.Pod, error)) (*corev1.Pod, error)
	EnsureService(name, namespace string, factory func() (*corev1.Service, error)) (*corev1.Service, error)
	EnsureConfigMap(name, namespace string, factory func() (*corev1.ConfigMap, error)) (*corev1.ConfigMap, error)
	EnsureJob(name, namespace string, factory func() (*v1.Job, error)) (*v1.Job, error)

	ApplyStatefulSet(name, namespace string, s *appsv1.StatefulSet) error
	ApplyService(name, namespace string, s *corev1.Service) error
	ApplyDeployment(name, namespace string, d *appsv1.Deployment) error
	ApplySecret(name, namespace string, s *corev1.Secret) error
	GetDeployment(name, namespace string) (*appsv1.Deployment, error)
	DeleteDeployment(name, namespace string) error
}

func New(kubeClientSet kubernetes.Interface, jobsLister batchlisters.JobLister) KubeClient {
	return &kubeclient{
		kubeClientSet: kubeClientSet,
		jobsLister:    jobsLister,
	}
}

type kubeclient struct {
	kubeClientSet kubernetes.Interface
	jobsLister    batchlisters.JobLister
}

func (c *kubeclient) EnsureDeployment(name, namespace string, factory func() (*appsv1.Deployment, error)) (*appsv1.Deployment, error) {
	deployment, err := c.kubeClientSet.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return deployment, err
	}
	deployment, err = factory()
	if err != nil {
		return nil, err
	}
	return c.kubeClientSet.AppsV1().Deployments(namespace).Create(deployment)
}

func (c *kubeclient) EnsurePod(name, namespace string, factory func() (*corev1.Pod, error)) (*corev1.Pod, error) {
	pod, err := c.kubeClientSet.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return pod, err
	}
	pod, err = factory()
	if err != nil {
		return nil, err
	}
	return c.kubeClientSet.CoreV1().Pods(namespace).Create(pod)
}

func (c *kubeclient) EnsureService(name, namespace string, factory func() (*corev1.Service, error)) (*corev1.Service, error) {
	service, err := c.kubeClientSet.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return service, err
	}
	service, err = factory()
	if err != nil {
		return nil, err
	}
	return c.kubeClientSet.CoreV1().Services(namespace).Create(service)
}

func (c *kubeclient) EnsureConfigMap(name, namespace string, factory func() (*corev1.ConfigMap, error)) (*corev1.ConfigMap, error) {
	configmap, err := c.kubeClientSet.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return configmap, err
	}
	configmap, err = factory()
	if err != nil {
		return nil, err
	}
	return c.kubeClientSet.CoreV1().ConfigMaps(namespace).Create(configmap)
}

func (c *kubeclient) EnsureJob(name, namespace string, factory func() (*v1.Job, error)) (*v1.Job, error) {
	job, err := c.jobsLister.Jobs(namespace).Get(name)
	if !errors.IsNotFound(err) {
		return job, err
	}
	job, err = factory()
	if err != nil {
		return nil, err
	}
	return c.kubeClientSet.BatchV1().Jobs(namespace).Create(job)
}

func (c *kubeclient) ApplyStatefulSet(name, namespace string, s *appsv1.StatefulSet) error {
	_, err := c.kubeClientSet.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = c.kubeClientSet.AppsV1().StatefulSets(namespace).Create(s)
		return err
	}
	if err == nil {
		_, err = c.kubeClientSet.AppsV1().StatefulSets(namespace).Update(s)
	}
	return err
}

func (c *kubeclient) ApplyService(name, namespace string, s *corev1.Service) error {
	_, err := c.kubeClientSet.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = c.kubeClientSet.CoreV1().Services(namespace).Create(s)
		return err
	}
	if err == nil {
		_, err = c.kubeClientSet.CoreV1().Services(namespace).Update(s)
	}
	return err
}

func (c *kubeclient) ApplyDeployment(name, namespace string, d *appsv1.Deployment) error {
	_, err := c.kubeClientSet.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = c.kubeClientSet.AppsV1().Deployments(namespace).Create(d)
		return err
	}
	if err == nil {
		_, err = c.kubeClientSet.AppsV1().Deployments(namespace).Update(d)
	}
	return err
}

func (c *kubeclient) ApplySecret(name, namespace string, s *corev1.Secret) error {
	_, err := c.kubeClientSet.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = c.kubeClientSet.CoreV1().Secrets(namespace).Create(s)
		return err
	}
	if err == nil {
		_, err = c.kubeClientSet.CoreV1().Secrets(namespace).Update(s)
	}
	return err
}

func (c *kubeclient) GetDeployment(name, namespace string) (*appsv1.Deployment, error) {
	return c.kubeClientSet.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

func (c *kubeclient) DeleteDeployment(name, namespace string) error {
	err := c.kubeClientSet.AppsV1().Deployments(namespace).Delete(name, nil)
	if err == nil || errors.IsNotFound(err) {
		return nil
	}
	return err
}
