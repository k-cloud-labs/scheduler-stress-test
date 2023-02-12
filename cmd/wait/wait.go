/*
Copyright © 2023 likakuli <1154584512@qq.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package wait

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	apiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

func NewCommand() *cobra.Command {
	opts := newOptions()

	cmd := &cobra.Command{
		Use:   "wait",
		Short: "to wait test result.",
		Long:  `to wait test result. for details, run help`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := run(apiserver.SetupSignalContext(), opts); err != nil {
				return err
			}
			return nil
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	fss := cliflag.NamedFlagSets{}

	genericFlagSet := fss.FlagSet("generic")
	opts.addFlags(genericFlagSet)

	// Set klog flags
	logsFlagSet := fss.FlagSet("logs")

	// Since klog only accepts golang flag set, so introduce a shim here.
	flagSetShim := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(flagSetShim)
	logsFlagSet.AddGoFlagSet(flagSetShim)

	cmd.Flags().AddFlagSet(genericFlagSet)
	cmd.Flags().AddFlagSet(logsFlagSet)

	return cmd
}

func run(ctx context.Context, opts *options) error {
	restConfig, err := clientcmd.BuildConfigFromFlags(opts.master, opts.kubeConfig)
	if err != nil {
		return err
	}
	restConfig.QPS, restConfig.Burst = opts.clusterAPIQPS, opts.clusterAPIBurst

	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0, informers.WithNamespace(opts.namespace))

	t := newTask(kubeClient, informerFactory, opts, ctx.Done())
	if err = t.Start(ctx); err != nil {
		klog.Errorf("task exits unexpectedly: %v", err)
		return err
	}
	return nil
}

type task struct {
	client          kubernetes.Interface
	informerFactory informers.SharedInformerFactory

	lister    corelister.PodLister
	namespace string
}

func (t *task) Start(ctx context.Context) error {
	stopCh := ctx.Done()
	klog.Infoln("Starting wait task")
	defer klog.Infoln("Stopping example task")

	t.informerFactory.Start(stopCh)
	t.informerFactory.WaitForCacheSync(stopCh)

	// add your logic here
	duration := 5 * time.Second
	ticker := time.Tick(duration)

	for {
		select {
		case <-ticker:
			pods, err := t.lister.Pods(t.namespace).List(labels.Everything())
			if err != nil {
				return fmt.Errorf("failed to list pods: %v", err)
			}

			var (
				scheduled int
				avg       time.Duration
				m         time.Duration
			)

			for _, pod := range pods {
				_, cond := getPodCondition(&pod.Status, corev1.PodScheduled)
				if cond != nil && cond.Status == corev1.ConditionTrue {
					scheduled++
					scheduleTime := cond.LastTransitionTime.Sub(pod.CreationTimestamp.Time)
					avg += scheduleTime
					if scheduleTime > m {
						m = scheduleTime
					}
				}
			}

			if len(pods) > 0 {
				avg /= time.Duration(len(pods))
			}

			klog.Infof("All: %d, Scheduled: %d, Unscheduled: %d, Max: %v, Avg: %v, Pods scheduler per second: %.2f",
				len(pods),
				scheduled,
				len(pods)-scheduled,
				m,
				avg,
				func() float64 {
					if m.Seconds() == 0 {
						return 0
					}
					return float64(len(pods)) / m.Seconds()
				}())

		case <-stopCh:
			break
		}
	}
}

func newTask(client kubernetes.Interface, factory informers.SharedInformerFactory, opts *options, done <-chan struct{}) *task {
	t := &task{
		client:          client,
		informerFactory: factory,

		lister:    factory.Core().V1().Pods().Lister(),
		namespace: opts.namespace,
	}

	// add your logic here
	return t
}

func getPodCondition(status *corev1.PodStatus, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return getPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func getPodConditionFromList(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}
