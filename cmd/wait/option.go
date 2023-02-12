package wait

import (
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type options struct {
	master     string
	kubeConfig string
	// ClusterAPIQPS is the QPS to use while talking with cluster kube-apiserver.
	clusterAPIQPS float32
	// ClusterAPIBurst is the burst to allow while talking with cluster kube-apiserver.
	clusterAPIBurst int
	namespace       string
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
	fs.Float32Var(&o.clusterAPIQPS, "kube-api-qps", 20.0, "QPS to use while talking with apiserver. Doesn't cover events and node heartbeat apis which rate limiting is controlled by a different set of flags.")
	fs.IntVar(&o.clusterAPIBurst, "kube-api-burst", 30, "Burst to use while talking with apiserver. Doesn't cover events and node heartbeat apis which rate limiting is controlled by a different set of flags.")
	fs.StringVar(&o.namespace, "namespace", metav1.NamespaceAll, "Namespace that take effect.")
}
