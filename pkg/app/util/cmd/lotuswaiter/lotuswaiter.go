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

package lotuswaiter

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	lotusv1beta1 "github.com/lotusload/lotus/pkg/app/lotus/apis/lotus/v1beta1"
	clientset "github.com/lotusload/lotus/pkg/app/lotus/client/clientset/versioned"
	"github.com/lotusload/lotus/pkg/cli"
)

type waiter struct {
	lotusName   string
	namespace   string
	phases      []string
	probePeriod time.Duration
	timeout     time.Duration
	kubeconfig  string
	masterURL   string
}

func NewCommand() *cobra.Command {
	w := &waiter{
		namespace: "default",
		phases: []string{
			string(lotusv1beta1.LotusFailed),
			string(lotusv1beta1.LotusSucceeded),
		},
		probePeriod: 30 * time.Second,
		timeout:     time.Hour,
	}
	cmd := &cobra.Command{
		Use:   "lotus-waiter",
		Short: "Wait until a give lotus reaches a phase",
		RunE:  cli.WithContext(w.run),
	}

	cmd.Flags().StringVar(&w.lotusName, "lotus-name", w.lotusName, "The name of waiting lotus")
	cmd.MarkFlagRequired("lotus-name")
	cmd.Flags().StringVar(&w.namespace, "namespace", w.namespace, "The namespace of waiting lotus")
	cmd.Flags().StringSliceVar(&w.phases, "phases", w.phases, "The list of waiting phases")
	cmd.Flags().DurationVar(&w.probePeriod, "probe-period", w.probePeriod, "How often to perform the probe")
	cmd.Flags().DurationVar(&w.timeout, "timeout", w.timeout, "The maximum waiting time")
	cmd.Flags().StringVar(&w.kubeconfig, "kube-config", w.kubeconfig, "Path to a kubeconfig. Only required if out-of-cluster.")
	cmd.Flags().StringVar(&w.masterURL, "master", w.masterURL, "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	return cmd
}

func (w *waiter) run(ctx context.Context, logger *zap.Logger) error {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()
	logger = logger.With(
		zap.String("lotus", w.lotusName),
		zap.String("namespace", w.namespace),
	)
	cfg, err := clientcmd.BuildConfigFromFlags(w.masterURL, w.kubeconfig)
	if err != nil {
		logger.Error("failed to build kube config", zap.Error(err))
		return err
	}
	lotusClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		logger.Error("failed to build lotus clientset", zap.Error(err))
		return err
	}
	return wait(ctx, w.probePeriod, w.phases, logger, func() (*lotusv1beta1.Lotus, error) {
		//TODO: I think using Get is simpler and better for our case. Consider about watch or informer later.
		return lotusClient.LotusV1beta1().Lotuses(w.namespace).Get(w.lotusName, metav1.GetOptions{})
	})
}
