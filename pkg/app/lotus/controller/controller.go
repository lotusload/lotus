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

package controller

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	batchinformers "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	lotusv1beta1 "github.com/lotusload/lotus/pkg/app/lotus/apis/lotus/v1beta1"
	clientset "github.com/lotusload/lotus/pkg/app/lotus/client/clientset/versioned"
	lotusscheme "github.com/lotusload/lotus/pkg/app/lotus/client/clientset/versioned/scheme"
	informers "github.com/lotusload/lotus/pkg/app/lotus/client/informers/externalversions/lotus/v1beta1"
	listers "github.com/lotusload/lotus/pkg/app/lotus/client/listers/lotus/v1beta1"
	"github.com/lotusload/lotus/pkg/app/lotus/config"
	"github.com/lotusload/lotus/pkg/app/lotus/kubeclient"
	"github.com/lotusload/lotus/pkg/app/lotus/model"
	"github.com/lotusload/lotus/pkg/app/lotus/resource"
)

type Controller struct {
	kubeClient     kubeclient.KubeClient
	lotusclientset clientset.Interface

	jobsSynced    cache.InformerSynced
	lotusesLister listers.LotusLister
	lotusesSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder

	namespace                string
	release                  string
	prometheusServiceAccount string
	configFile               string
	logger                   *zap.Logger
}

func NewController(
	kubeclientset kubernetes.Interface,
	lotusclientset clientset.Interface,
	jobInformer batchinformers.JobInformer,
	lotusInformer informers.LotusInformer,
	namespace string,
	release string,
	prometheusServiceAccount string,
	configFile string,
	logger *zap.Logger) *Controller {

	logger = logger.Named("controller")
	logger.Info("creating event broadcaster")
	lotusscheme.AddToScheme(scheme.Scheme)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.Sugar().Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: kubeclientset.CoreV1().Events(""),
	})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{
		Component: "lotus-controller",
	})

	controller := &Controller{
		kubeClient:               kubeclient.New(kubeclientset, jobInformer.Lister()),
		lotusclientset:           lotusclientset,
		jobsSynced:               jobInformer.Informer().HasSynced,
		lotusesLister:            lotusInformer.Lister(),
		lotusesSynced:            lotusInformer.Informer().HasSynced,
		workqueue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Lotuses"),
		recorder:                 recorder,
		namespace:                namespace,
		release:                  release,
		prometheusServiceAccount: prometheusServiceAccount,
		configFile:               configFile,
		logger:                   logger,
	}
	lotusInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueLotus,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueLotus(new)
		},
	})
	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			controller.onObject(new)
		},
	})
	return controller
}

func (c *Controller) Run(ctx context.Context, workers int) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.logger.Info("starting Lotus controller")
	c.logger.Info("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.jobsSynced, c.lotusesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	// Update static resources based on new configuration
	c.logger.Info("updating static resources")
	if err := c.ensureStaticResources(); err != nil {
		c.logger.Error("failed to ensure static resources", zap.Error(err))
		return err
	}

	c.logger.Info("informer caches synced")
	c.logger.Info("starting workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, ctx.Done())
	}

	c.logger.Info("started workers", zap.Int("workers", workers))
	<-ctx.Done()
	c.logger.Info("shutting down workers")
	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		key, ok := obj.(string)
		if !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		c.logger.Info("successfully synced item", zap.String("key", key))
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

// Compares the actual state with the desired and attempts to converge the two.
// It then updates the Status block of the Lotus resource with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	lotus, err := c.lotusesLister.Lotuses(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("lotus '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	switch lotus.Status.Phase {
	case lotusv1beta1.LotusInit:
		return c.updateLotusStatus(lotus, lotusv1beta1.LotusPending)
	case lotusv1beta1.LotusPending:
		return c.updateLotusStatus(lotus, lotusv1beta1.LotusPreparing)
	case lotusv1beta1.LotusPreparing:
		if lotus.Spec.Preparer == nil {
			return c.toRunningPhase(lotus)
		}
		return c.syncPreparingLotus(lotus)
	case lotusv1beta1.LotusRunning:
		return c.syncRunningLotus(lotus)
	case lotusv1beta1.LotusCleaning:
		if lotus.Spec.Cleaner == nil {
			return c.updateLotusStatus(lotus, lotusv1beta1.LotusSucceeded)
		}
		return c.syncCleaningLotus(lotus)
	case lotusv1beta1.LotusFailureCleaning:
		if lotus.Spec.Cleaner == nil {
			return c.updateLotusStatus(lotus, lotusv1beta1.LotusFailed)
		}
		return c.syncFailureCleaningLotus(lotus)
	case lotusv1beta1.LotusSucceeded:
		return nil
	case lotusv1beta1.LotusFailed:
		return nil
	}
	c.logger.Warn("unexpected lotus phase", zap.String("phase", string(lotus.Status.Phase)))
	return nil
}

func (c *Controller) syncPreparingLotus(lotus *lotusv1beta1.Lotus) error {
	factory := resource.NewFactory(lotus, c.configFile)
	jobName := factory.PreparerJobName()
	job, err := c.kubeClient.EnsureJob(jobName, lotus.Namespace, factory.NewPreparerJob)
	if err != nil {
		return err
	}
	if job.Status.Failed > 0 {
		return c.updateLotusStatus(lotus, lotusv1beta1.LotusFailureCleaning)
	}
	if job.Status.Succeeded > 0 {
		return c.toRunningPhase(lotus)
	}
	c.logger.Info("preparer job is still running", zap.String("name", jobName))
	return nil
}

func (c *Controller) toRunningPhase(lotus *lotusv1beta1.Lotus) error {
	if err := c.ensurePrometheusResources(lotus); err != nil {
		return err
	}
	if err := c.ensureWorkerResources(lotus); err != nil {
		return err
	}
	factory := resource.NewFactory(lotus, c.configFile)
	name := factory.MonitorJobName()
	if _, err := c.kubeClient.EnsureConfigMap(name, lotus.Namespace, factory.NewMonitorConfigMap); err != nil {
		return err
	}
	return c.updateLotusStatus(lotus, lotusv1beta1.LotusRunning)
}

func (c *Controller) syncRunningLotus(lotus *lotusv1beta1.Lotus) error {
	factory := resource.NewFactory(lotus, c.configFile)
	jobName := factory.MonitorJobName()
	job, err := c.kubeClient.EnsureJob(jobName, lotus.Namespace, factory.NewMonitorJob)
	if err != nil {
		return err
	}
	if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		c.logger.Info("monitor job is still running", zap.String("name", jobName))
		return nil
	}
	// Scale down or Delete worker deployment.
	workerName := factory.WorkerName()
	err = c.kubeClient.DeleteDeployment(workerName, lotus.Namespace)
	if err != nil {
		c.logger.Error("failed to delete worker deployment", zap.Error(err))
		return err
	}
	if job.Status.Failed > 0 {
		return c.updateLotusStatus(lotus, lotusv1beta1.LotusFailureCleaning)
	}
	return c.updateLotusStatus(lotus, lotusv1beta1.LotusCleaning)
}

func (c *Controller) syncCleaningLotus(lotus *lotusv1beta1.Lotus) error {
	factory := resource.NewFactory(lotus, c.configFile)
	jobName := factory.CleanerJobName()
	job, err := c.kubeClient.EnsureJob(jobName, lotus.Namespace, factory.NewCleanerJob)
	if err != nil {
		return err
	}
	if job.Status.Succeeded > 0 {
		return c.updateLotusStatus(lotus, lotusv1beta1.LotusSucceeded)
	}
	if job.Status.Failed > 0 {
		return c.updateLotusStatus(lotus, lotusv1beta1.LotusFailed)
	}
	return nil
}

func (c *Controller) syncFailureCleaningLotus(lotus *lotusv1beta1.Lotus) error {
	factory := resource.NewFactory(lotus, c.configFile)
	jobName := factory.CleanerJobName()
	job, err := c.kubeClient.EnsureJob(jobName, lotus.Namespace, factory.NewCleanerJob)
	if err != nil {
		return err
	}
	if job.Status.Succeeded > 0 || job.Status.Failed > 0 {
		return c.updateLotusStatus(lotus, lotusv1beta1.LotusFailed)
	}
	return nil
}

func (c *Controller) ensureWorkerResources(lotus *lotusv1beta1.Lotus) error {
	factory := resource.NewFactory(lotus, c.configFile)
	name := factory.WorkerName()
	if _, err := c.kubeClient.EnsureService(name, lotus.Namespace, factory.NewWorkerService); err != nil {
		return err
	}
	_, err := c.kubeClient.EnsureDeployment(name, lotus.Namespace, factory.NewWorkerDeployment)
	return err
}

func (c *Controller) ensurePrometheusResources(lotus *lotusv1beta1.Lotus) error {
	factory := resource.NewFactory(lotus, c.configFile)
	name := factory.PrometheusName()
	if _, err := c.kubeClient.EnsureConfigMap(name, lotus.Namespace, factory.NewPrometheusConfigMap); err != nil {
		return err
	}
	podFactory := func() (*corev1.Pod, error) {
		return factory.NewPrometheusPod(c.prometheusServiceAccount, c.release)
	}
	if _, err := c.kubeClient.EnsurePod(name, lotus.Namespace, podFactory); err != nil {
		return err
	}
	_, err := c.kubeClient.EnsureService(name, lotus.Namespace, factory.NewPrometheusService)
	return err
}

func (c *Controller) enqueueLotus(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.logger.Info("enqueue a lotus", zap.String("key", key))
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) onObject(obj interface{}) {
	object, ok := obj.(metav1.Object)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok := tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		c.logger.Info("recovered deleted object from tombstone", zap.String("name", object.GetName()))
	}
	ownerRef := metav1.GetControllerOf(object)
	if ownerRef == nil {
		return
	}
	if ownerRef.Kind != model.LotusKind {
		return
	}
	lotus, err := c.lotusesLister.Lotuses(object.GetNamespace()).Get(ownerRef.Name)
	if err != nil {
		c.logger.Info("ignoring orphaned object",
			zap.String("object_self_link", object.GetSelfLink()),
			zap.String("lotus_name", ownerRef.Name))
		return
	}
	c.logger.Info("will enqueue new lotus because of an object change",
		zap.String("object_self_link", object.GetSelfLink()),
		zap.String("object_name", object.GetName()))
	c.enqueueLotus(lotus)
}

func (c *Controller) updateLotusStatus(lotus *lotusv1beta1.Lotus, phase lotusv1beta1.LotusPhase) error {
	lotus = copyWithNewStatus(lotus, phase)
	_, err := c.lotusclientset.LotusV1beta1().Lotuses(lotus.Namespace).Update(lotus)
	return err
}

func copyWithNewStatus(lotus *lotusv1beta1.Lotus, phase lotusv1beta1.LotusPhase) *lotusv1beta1.Lotus {
	lotusCopy := lotus.DeepCopy()
	prev := lotusCopy.Status.Phase
	if prev != phase {
		now := metav1.Now()
		switch phase {
		case lotusv1beta1.LotusPreparing:
			lotusCopy.Status.PreparerStartTime = &now
		case lotusv1beta1.LotusRunning:
			lotusCopy.Status.WorkerStartTime = &now
		case lotusv1beta1.LotusCleaning:
			fallthrough
		case lotusv1beta1.LotusFailureCleaning:
			lotusCopy.Status.CleanerStartTime = &now
		}
		switch prev {
		case lotusv1beta1.LotusPreparing:
			lotusCopy.Status.PreparerCompletionTime = &now
		case lotusv1beta1.LotusRunning:
			lotusCopy.Status.WorkerCompletionTime = &now
		case lotusv1beta1.LotusCleaning:
			fallthrough
		case lotusv1beta1.LotusFailureCleaning:
			lotusCopy.Status.CleanerCompletionTime = &now
		}
	}
	lotusCopy.Status.Phase = phase
	return lotusCopy
}

func (c *Controller) ensureStaticResources() error {
	controllerDeployment, err := c.kubeClient.GetDeployment("lotus-controller", c.namespace)
	if err != nil {
		c.logger.Error("failed to get controller deployment", zap.Error(err))
		return err
	}
	owners := []metav1.OwnerReference{
		*metav1.NewControllerRef(controllerDeployment, schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    "Deployment",
		}),
	}

	f := resource.NewStaticResourceFactory(c.namespace, c.release, c.configFile, owners)
	thanosPeerService, err := f.NewThanosPeerService()
	if err != nil {
		return err
	}
	if err := c.kubeClient.ApplyService(f.ThanosPeerName(), c.namespace, thanosPeerService); err != nil {
		return err
	}

	cfg, err := config.FromFile(c.configFile)
	if err != nil {
		return err
	}
	if cfg.TimeSeriesStorage != nil {
		timeSeriesStoreSecret, err := f.NewTimeSeriesStoreConfigSecret()
		if err != nil {
			return err
		}
		if err := c.kubeClient.ApplySecret(f.TimeSeriesStoreConfigSecretName(), c.namespace, timeSeriesStoreSecret); err != nil {
			return err
		}
		thanosStore, err := f.NewThanosStoreStatefulSet()
		if err != nil {
			return err
		}
		if err := c.kubeClient.ApplyStatefulSet(f.ThanosStoreName(), c.namespace, thanosStore); err != nil {
			return err
		}
	}

	thanosQueryDeployment, err := f.NewThanosQueryDeployment()
	if err != nil {
		return err
	}
	if err := c.kubeClient.ApplyDeployment(f.ThanosQueryName(), c.namespace, thanosQueryDeployment); err != nil {
		return err
	}
	thanosQueryService, err := f.NewThanosQueryService()
	if err != nil {
		return err
	}
	return c.kubeClient.ApplyService(f.ThanosQueryName(), c.namespace, thanosQueryService)
}
