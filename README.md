# scheduler-stress-test
This is a stress testing tool for the scheduler in a large-scale scenario.

### Preparatory Work
To simulate a large-scale scheduling scenario, you can create a large number of nodes as needed using **kwok**. The nodes you create may be in a **NotReady** state. In order to use the nodes to schedule pods, the "pod.yaml" file must add a tolerance to tolerate all **NoSchedule** taints.

To prepare for this, you should do the following:

1. Install **kwok** on your k8s cluster. Refer to https://kwok.sigs.k8s.io/docs/user/kwok-in-cluster/.
2. Create fake nodes on your k8s cluster.

fake node example:
```shell
cat << EOF > node.yaml 
apiVersion: v1
kind: Node
metadata:
  annotations:
    node.alpha.kubernetes.io/ttl: "0"
    kwok.x-k8s.io/node: fake
  labels:
    beta.kubernetes.io/arch: amd64
    beta.kubernetes.io/os: linux
    kubernetes.io/arch: amd64
    kubernetes.io/hostname: {NODE_NAME}
    kubernetes.io/os: linux
    kubernetes.io/role: agent
    node-role.kubernetes.io/agent: ""
    type: kwok
  name: {NODE_NAME}
spec:
  taints: # Avoid scheduling actual running pods to fake Node
    - effect: NoSchedule
      key: kwok.x-k8s.io/node
      value: fake
status:
  allocatable:
    cpu: "64"
    ephemeral-storage: "289839513121"
    hugepages-1Gi: "0"
    hugepages-2Mi: "0"
    memory: 250Gi
    pods: "110"
  capacity:
    cpu: "64"
    ephemeral-storage: 307125Mi
    hugepages-1Gi: "0"
    hugepages-2Mi: "0"
    memory: 250Gi
    pods: "128"
  nodeInfo:
    architecture: amd64
    bootID: ""
    containerRuntimeVersion: ""
    kernelVersion: ""
    kubeProxyVersion: fake
    kubeletVersion: fake
    machineID: ""
    operatingSystem: linux
    osImage: ""
    systemUUID: ""
  phase: Running
EOF

# create nodes as you needed
for i in {0..99}; do sed "s/{NODE_NAME}/kwok-node-$i/g" node.yaml | kubectl apply -f -; done
```


## How to use
Download the code and build it:
```shell
git clone https://github.com/k-cloud-labs/scheduler-stress-test.git
make build
```
Two commands are supported: create and wait.  

The **create** command creates a specified number of pods from a template file in the k8s cluster with a specified level of concurrency.  

The **wait** command waits for all the pods created above to be scheduled and continuously prints the results.

Example:
```shell
# create 1000 pod by 1000 concurrency (namespace: scheduler-stress-test)
sst create --kubeconfig=/root/.kube/config --count 1000 --concurrency 1000 --pod-template=pod.yaml
# wait result
sst wait --kubeconfig=/root/.kube/config --namespace=seduler-stress-test
```

The example above uses "pod.yaml" as a template and creates 1000 pods in the "scheduler-stress-test" namespace of the k8s cluster. It then waits and continuously prints the results. You can modify the "pod.yaml" file as needed.  
