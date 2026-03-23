package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vasudevchavan/kubecaps/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes clientset and provides typed accessors.
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
	dynClient dynamicClient
}

// NewClient creates a new Kubernetes client. It auto-detects kubeconfig from:
// 1. Explicit path (if provided)
// 2. KUBECONFIG env variable
// 3. Default ~/.kube/config
// 4. In-cluster config
func NewClient(kubeconfigPath string) (*Client, error) {
	config, err := buildConfig(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// buildConfig resolves kubeconfig from path, env, default, or in-cluster.
func buildConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		return clientcmd.BuildConfigFromFlags("", envPath)
	}

	home, err := os.UserHomeDir()
	if err == nil {
		defaultPath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(defaultPath); err == nil {
			return clientcmd.BuildConfigFromFlags("", defaultPath)
		}
	}

	return rest.InClusterConfig()
}

// Clientset returns the underlying Kubernetes clientset.
func (c *Client) Clientset() *kubernetes.Clientset {
	return c.clientset
}

// ListDeployments lists all Deployments in the given namespace.
func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]types.WorkloadInfo, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments in namespace %s: %w", namespace, err)
	}

	var workloads []types.WorkloadInfo
	for _, d := range deployments.Items {
		workloads = append(workloads, deploymentToWorkloadInfo(&d))
	}
	return workloads, nil
}

// ListStatefulSets lists all StatefulSets in the given namespace.
func (c *Client) ListStatefulSets(ctx context.Context, namespace string) ([]types.WorkloadInfo, error) {
	statefulsets, err := c.clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulsets in namespace %s: %w", namespace, err)
	}

	var workloads []types.WorkloadInfo
	for _, s := range statefulsets.Items {
		workloads = append(workloads, statefulSetToWorkloadInfo(&s))
	}
	return workloads, nil
}

// ListDaemonSets lists all DaemonSets in the given namespace.
func (c *Client) ListDaemonSets(ctx context.Context, namespace string) ([]types.WorkloadInfo, error) {
	daemonsets, err := c.clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list daemonsets in namespace %s: %w", namespace, err)
	}

	var workloads []types.WorkloadInfo
	for _, d := range daemonsets.Items {
		workloads = append(workloads, daemonSetToWorkloadInfo(&d))
	}
	return workloads, nil
}

// ListAllWorkloads lists all Deployments, StatefulSets, and DaemonSets in a namespace.
func (c *Client) ListAllWorkloads(ctx context.Context, namespace string) ([]types.WorkloadInfo, error) {
	var allWorkloads []types.WorkloadInfo

	deployments, err := c.ListDeployments(ctx, namespace)
	if err != nil {
		return nil, err
	}
	allWorkloads = append(allWorkloads, deployments...)

	statefulsets, err := c.ListStatefulSets(ctx, namespace)
	if err != nil {
		return nil, err
	}
	allWorkloads = append(allWorkloads, statefulsets...)

	daemonsets, err := c.ListDaemonSets(ctx, namespace)
	if err != nil {
		return nil, err
	}
	allWorkloads = append(allWorkloads, daemonsets...)

	return allWorkloads, nil
}

// GetDeployment returns a specific Deployment as a WorkloadInfo.
func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*types.WorkloadInfo, error) {
	d, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	w := deploymentToWorkloadInfo(d)
	return &w, nil
}

// GetStatefulSet returns a specific StatefulSet as a WorkloadInfo.
func (c *Client) GetStatefulSet(ctx context.Context, namespace, name string) (*types.WorkloadInfo, error) {
	s, err := c.clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	w := statefulSetToWorkloadInfo(s)
	return &w, nil
}

// GetDaemonSet returns a specific DaemonSet as a WorkloadInfo.
func (c *Client) GetDaemonSet(ctx context.Context, namespace, name string) (*types.WorkloadInfo, error) {
	d, err := c.clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	w := daemonSetToWorkloadInfo(d)
	return &w, nil
}

// GetWorkload tries to find a workload by name across Deployments, StatefulSets, and DaemonSets.
// It also handles Pod names by resolving them to their owning workload.
func (c *Client) GetWorkload(ctx context.Context, namespace, name string) (*types.WorkloadInfo, error) {
	// 1. Try direct match across workload types
	if w, err := c.GetDeployment(ctx, namespace, name); err == nil {
		return w, nil
	}
	if w, err := c.GetStatefulSet(ctx, namespace, name); err == nil {
		return w, nil
	}
	if w, err := c.GetDaemonSet(ctx, namespace, name); err == nil {
		return w, nil
	}

	// 2. Try as a Pod name and resolve owner
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "StatefulSet" {
				return c.GetStatefulSet(ctx, namespace, owner.Name)
			}
			if owner.Kind == "ReplicaSet" {
				// Resolve ReplicaSet to Deployment
				rs, err := c.clientset.AppsV1().ReplicaSets(namespace).Get(ctx, owner.Name, metav1.GetOptions{})
				if err == nil {
					for _, rsOwner := range rs.OwnerReferences {
						if rsOwner.Kind == "Deployment" {
							return c.GetDeployment(ctx, namespace, rsOwner.Name)
						}
					}
				}
			}
			if owner.Kind == "DaemonSet" {
				return c.GetDaemonSet(ctx, namespace, owner.Name)
			}
		}
	}

	return nil, fmt.Errorf("workload %s not found in namespace %s (tried Deployment, StatefulSet, DaemonSet, and Pod owner resolution)", name, namespace)
}

func deploymentToWorkloadInfo(d *appsv1.Deployment) types.WorkloadInfo {
	replicas := int32(1)
	if d.Spec.Replicas != nil {
		replicas = *d.Spec.Replicas
	}
	return types.WorkloadInfo{
		Name:      d.Name,
		Namespace: d.Namespace,
		Kind:      "Deployment",
		Labels:    d.Labels,
		Replicas:  replicas,
	}
}

func statefulSetToWorkloadInfo(s *appsv1.StatefulSet) types.WorkloadInfo {
	replicas := int32(1)
	if s.Spec.Replicas != nil {
		replicas = *s.Spec.Replicas
	}
	return types.WorkloadInfo{
		Name:      s.Name,
		Namespace: s.Namespace,
		Kind:      "StatefulSet",
		Labels:    s.Labels,
		Replicas:  replicas,
	}
}

func daemonSetToWorkloadInfo(d *appsv1.DaemonSet) types.WorkloadInfo {
	return types.WorkloadInfo{
		Name:      d.Name,
		Namespace: d.Namespace,
		Kind:      "DaemonSet",
		Labels:    d.Labels,
		Replicas:  d.Status.DesiredNumberScheduled,
	}
}
