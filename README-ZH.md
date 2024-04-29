[English](./README.md)
# scheduler-stress-test
这是一个用于模拟大规模场景下调度器压力测试的工具。

## 环境准备

为了模拟大规模调度场景，您可以使用 kwok 创建所需数量的节点。创建的节点可能处于 `NotReady` 状态。为了能够将这些节点用于调度 `Pod`，必须为待调度的 `Pod` 添加一个 `toleration`，以容忍所有 `NoSchedule` 的污点。

为此，您应该执行以下步骤：

1. 在您的 k8s 集群上安装 `kwok`，请参考 https://kwok.sigs.k8s.io/docs/user/kwok-in-cluster/；

2. 在您的 k8s 集群上创建虚拟节点，可以参考如下命令

   ```bash
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
       ephemeral-storage: 1Ti
       hugepages-1Gi: "0"
       hugepages-2Mi: "0"
       memory: 250Gi
       pods: "110"
     capacity:
       cpu: "64"
       ephemeral-storage: 1Ti
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

## 压测

下载代码并构建：

```bash
git clone https://github.com/k-cloud-labs/scheduler-stress-test.git 

make build
```

该工具支持两个命令：create 和 wait。

create 命令使用指定的模板文件，在 k8s 集群中以指定的并发级别创建指定数量的 pod。

wait 命令等待所有上述创建的 pod 被调度并连续打印结果。

示例：

```bash
# 创建 1000 个 pod，使用 1000 的并发级别（namespace: scheduler-stress-test）
sst create --kubeconfig=/root/.kube/config --count 1000 --concurrency 1000 --pod-template=pod.yaml

# 等待结果
sst wait --kubeconfig=/root/.kube/config --namespace=scheduler-stress-test
```

上述示例使用项目中的 pod.yaml 作为模板，在 k8s 集群的 scheduler-stress-test 命名空间中创建了 1000 个 pod。然后等待并连续打印结果，您可以根据需要修改 pod.yaml 文件。

