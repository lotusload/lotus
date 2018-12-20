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
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/lotusload/lotus/pkg/app/lotus/client/clientset/versioned"
	informers "github.com/lotusload/lotus/pkg/app/lotus/client/informers/externalversions"
	lotus "github.com/lotusload/lotus/pkg/app/lotus/controller"
	"github.com/lotusload/lotus/pkg/cli"
)

type controller struct {
	kubeconfig               string
	masterURL                string
	namespace                string
	release                  string
	prometheusServiceAccount string
	configFile               string
}

func NewCommand() *cobra.Command {
	c := &controller{
		namespace: "default",
		release:   "lotus",
	}
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Start running Lotus controller",
		RunE:  cli.WithContext(c.run),
	}
	cmd.Flags().StringVar(&c.kubeconfig, "kube-config", c.kubeconfig, "Path to a kubeconfig. Only required if out-of-cluster.")
	cmd.Flags().StringVar(&c.masterURL, "master", c.masterURL, "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	cmd.Flags().StringVar(&c.namespace, "namespace", c.namespace, "The namespace of controller.")
	cmd.Flags().StringVar(&c.release, "release", c.release, "The release name of deployment.")
	cmd.Flags().StringVar(&c.prometheusServiceAccount, "prometheus-service-account", c.prometheusServiceAccount, "The name of service account for prometheus pods. This is required when rbac is enabled.")
	cmd.Flags().StringVar(&c.configFile, "config-file", c.configFile, "Path to the configuration file.")
	cmd.MarkFlagRequired("config-file")
	return cmd
}

func (c *controller) run(ctx context.Context, logger *zap.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags(c.masterURL, c.kubeconfig)
	if err != nil {
		logger.Error("failed to build kube config", zap.Error(err))
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Error("failed to build kubernetes clientset", zap.Error(err))
		return err
	}

	lotusClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		logger.Error("failed to build lotus clientset", zap.Error(err))
		return err
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(
		kubeClient,
		30*time.Second,
		kubeinformers.WithNamespace(c.namespace),
	)

	lotusInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		lotusClient,
		30*time.Second,
		informers.WithNamespace(c.namespace),
	)

	controller := lotus.NewController(
		kubeClient,
		lotusClient,
		kubeInformerFactory.Batch().V1().Jobs(),
		lotusInformerFactory.Lotus().V1beta1().Lotuses(),
		c.namespace,
		c.release,
		c.prometheusServiceAccount,
		c.configFile,
		logger,
	)

	kubeInformerFactory.Start(ctx.Done())
	lotusInformerFactory.Start(ctx.Done())

	if err = controller.Run(ctx, 1); err != nil {
		logger.Error("failed to run controller", zap.Error(err))
		return err
	}

	<-ctx.Done()
	return nil
}
