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
package create

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

// NewCommand represents the create command
func NewCommand() *cobra.Command {
	opts := newOptions()

	cmd := &cobra.Command{
		Use:   "create",
		Short: "to create pod concurrency",
		Long:  `to create pod concurrency. for details, run help.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(opts)
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

func run(opts *options) error {
	restConfig, err := clientcmd.BuildConfigFromFlags(opts.master, opts.kubeConfig)
	if err != nil {
		return err
	}
	restConfig.QPS, restConfig.Burst = float32(opts.concurrency), opts.concurrency*2

	kubeClient := kubernetes.NewForConfigOrDie(restConfig)

	pod, err := parsePodTemplate(opts.podTemplate)
	if err != nil {
		return err
	}

	start := time.Now()
	err = createPodConcurrent(kubeClient, opts.count, opts.concurrency, pod)
	end := time.Now()

	if err != nil {
		return nil
	}

	klog.Infof("create %d pods in %d milliseconds by %d concurrency.", opts.count, end.Sub(start).Milliseconds(), opts.concurrency)

	return nil
}

func createPodConcurrent(kubeClient *kubernetes.Clientset, count, concurrent int, pod *v1.Pod) error {
	name := strings.TrimRight(pod.Name, "-") + "-"

	podCh := make(chan *v1.Pod, count)
	for i := 0; i < count; i++ {
		copyPod := pod.DeepCopy()
		copyPod.Name = name + strconv.Itoa(i)
		podCh <- copyPod
	}

	stopCh := make(chan struct{})
	consumed := make(chan struct{}, count)
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				klog.Infof("%d pod(s) have been created.", len(consumed))
				if len(podCh) == 0 && len(consumed) == count {
					close(stopCh)
				}
			}
		}
	}()

	for i := 0; i < concurrent; i++ {
		go func() {
			for {
				select {
				case pod := <-podCh:
					_, err := kubeClient.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
					if err != nil {
						klog.Errorf("failed to create pod %s, err: %s", pod.Namespace+"/"+pod.Name, err.Error())
						podCh <- pod
					} else {
						consumed <- struct{}{}
					}
				}
			}
		}()
	}

	<-stopCh
	return nil
}

func parsePodTemplate(template string) (*v1.Pod, error) {
	filename, _ := filepath.Abs(template)
	spec, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open pod template file: %v", err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(spec, 4096)
	pod := &v1.Pod{}
	err = decoder.Decode(pod)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pod template file: %v", err)
	}

	return pod, nil
}
