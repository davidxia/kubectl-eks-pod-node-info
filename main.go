package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	selector string
	output   string
)

func main() {
	configFlags := genericclioptions.NewConfigFlags(true)

	cmd := &cobra.Command{
		Use:   "kubectl-eks-pod-node-info [flags] [pod...]",
		Short: "Show EC2 instance ID and type for EKS nodes running pods",
		Long: `eks-pod-node-info shows the EC2 instance ID and instance type for the
nodes running the specified pods in an EKS cluster.

Examples:
  # Show node info for pods matching a label selector
  kubectl eks-pod-node-info -n input-plane -l app.kubernetes.io/name=gateway

  # Show node info for specific pods
  kubectl eks-pod-node-info -n default my-pod-abc my-pod-def

  # Use a specific context
  kubectl eks-pod-node-info --context my-eks-context -n kube-system -l k8s-app=kube-dns

  # Show availability zone as an additional column
  kubectl eks-pod-node-info -o wide -n default my-pod-abc`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(configFlags, selector, output, args)
		},
	}

	configFlags.AddFlags(cmd.Flags())
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "Selector (label query) to filter pods")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output format. Use 'wide' to show additional columns including availability zone")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type podEntry struct {
	podName  string
	nodeName string
}

type nodeInfo struct {
	instanceID   string
	instanceType string
	zone         string
}

func run(configFlags *genericclioptions.ConfigFlags, selector, output string, podNames []string) error {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("building REST config: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}

	namespace, _, err := configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return fmt.Errorf("resolving namespace: %w", err)
	}

	var pods []podEntry

	if len(podNames) > 0 {
		for _, name := range podNames {
			pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("getting pod %q: %w", name, err)
			}
			pods = append(pods, podEntry{podName: pod.Name, nodeName: pod.Spec.NodeName})
		}
	} else {
		podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return fmt.Errorf("listing pods: %w", err)
		}
		if len(podList.Items) == 0 {
			fmt.Fprintln(os.Stderr, "No pods found")
			return nil
		}
		for i := range podList.Items {
			p := &podList.Items[i]
			pods = append(pods, podEntry{podName: p.Name, nodeName: p.Spec.NodeName})
		}
	}

	nodeCache := map[string]nodeInfo{}
	for _, p := range pods {
		if p.nodeName == "" {
			continue
		}
		if _, cached := nodeCache[p.nodeName]; cached {
			continue
		}
		node, err := client.CoreV1().Nodes().Get(context.TODO(), p.nodeName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("getting node %q: %w", p.nodeName, err)
		}
		instanceID := ""
		if id := node.Spec.ProviderID; id != "" {
			parts := strings.Split(id, "/")
			instanceID = parts[len(parts)-1]
		}
		nodeCache[p.nodeName] = nodeInfo{
			instanceID:   instanceID,
			instanceType: node.Labels["node.kubernetes.io/instance-type"],
			zone:         node.Labels["topology.kubernetes.io/zone"],
		}
	}

	wide := output == "wide"
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if wide {
		fmt.Fprintln(w, "POD\tNODE\tINSTANCE-ID\tINSTANCE-TYPE\tAVAILABILITY-ZONE")
	} else {
		fmt.Fprintln(w, "POD\tNODE\tINSTANCE-ID\tINSTANCE-TYPE")
	}
	for _, p := range pods {
		info := nodeCache[p.nodeName]
		if wide {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", p.podName, p.nodeName, info.instanceID, info.instanceType, info.zone)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.podName, p.nodeName, info.instanceID, info.instanceType)
		}
	}
	return w.Flush()
}
