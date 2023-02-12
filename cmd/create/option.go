package create

import (
	"github.com/spf13/pflag"
)

type options struct {
	master      string
	kubeConfig  string
	podTemplate string
	count       int
	concurrency int
}

func newOptions() *options {
	return &options{}
}

// addFlags adds flags of scheduler to the specified FlagSet
func (o *options) addFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.StringVar(&o.kubeConfig, "kubeconfig", o.kubeConfig, "Path to control plane kubeconfig file.")
	fs.StringVar(&o.master, "master", o.master, "The address of the member Kubernetes API server. Overrides any value in KubeConfig. Only required if out-of-cluster.")
	fs.StringVar(&o.podTemplate, "pod-template", o.podTemplate, "Path to pod template.")
	fs.IntVar(&o.count, "count", o.count, "Count of pod to create.")
	fs.IntVar(&o.concurrency, "concurrency", o.concurrency, "Concurrency of pod to create.")
}
