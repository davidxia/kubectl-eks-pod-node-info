# eks-pod-node-info

A `kubectl` plugin that shows the EC2 info for the nodes running your EKS pods.

## Installation

### Via krew

```sh
kubectl krew install eks-pod-node-info
```

### Manual

Download a release binary for your platform from the [releases page](https://github.com/davidxia/kubectl-eks-pod-node-info/releases),
rename it to `kubectl-eks-pod-node-info`, make it executable, and place it on your `$PATH`.

## Usage

```sh
kubectl eks-pod-node-info [flags] [pod...]

Flags:
  -l, --selector string    Selector (label query) to filter pods
  -n, --namespace string   Namespace of the pods
  -o, --output string      Output format. Use 'wide' to show additional columns including availability zone
      --context string     Kubernetes context to use
      --kubeconfig string  Path to kubeconfig file
  -h, --help               Show help
```

### Examples

```sh
# Pods matching a label selector
kubectl eks-pod-node-info -n input-plane -l app.kubernetes.io/name=input-plane-gateway

# Specific pods by name
kubectl eks-pod-node-info -n default my-pod-abc my-pod-def

# Use a non-default context
kubectl eks-pod-node-info --context eks-dev-us-east-1 -n kube-system -l k8s-app=kube-dns

# Show availability zone column
kubectl eks-pod-node-info -n input-plane -l app.kubernetes.io/name=input-plane-gateway -o wide 
```

### Sample output

```text
POD                                    NODE                             INSTANCE-ID            INSTANCE-TYPE
input-plane-gateway-5d7f9b8d6-xk2pq    ip-10-0-1-5.us-east-1.compute    i-0abc123def456789     m5.2xlarge
input-plane-gateway-5d7f9b8d6-zr9qm    ip-10-0-2-8.us-east-1.compute    i-0def456abc123789     m5.2xlarge
```

With `-o wide`:

```text
POD                                   NODE                           INSTANCE-ID            INSTANCE-TYPE   AVAILABILITY-ZONE
input-plane-gateway-5d7f9b8d6-xk2pq   ip-10-0-1-5.us-east-1.compute  i-0abc123def456789     m5.2xlarge      us-east-1a
input-plane-gateway-5d7f9b8d6-zr9qm   ip-10-0-2-8.us-east-1.compute  i-0def456abc123789     m5.2xlarge      us-east-1b
```

## Building from source

Requires Go 1.21+.

```sh
git clone https://github.com/davidxia/kubectl-eks-pod-node-info
cd kubectl-eks-pod-node-info
go build -o kubectl-eks-pod-node-info .
```
