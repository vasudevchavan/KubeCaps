package k8s

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// dynamicClient wraps the Kubernetes dynamic client for CRD access.
type dynamicClient = dynamic.Interface

// newDynamicClient creates a new dynamic Kubernetes client.
func newDynamicClient(config *rest.Config) (dynamic.Interface, error) {
	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	return dc, nil
}
