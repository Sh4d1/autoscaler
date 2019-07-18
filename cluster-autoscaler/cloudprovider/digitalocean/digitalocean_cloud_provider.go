/*
Copyright 2019 The Kubernetes Authors.

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

package digitalocean

import (
	"fmt"
	"io"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog"
)

var _ cloudprovider.CloudProvider = (*digitaloceanCloudProvider)(nil)

const (
	// ProviderName is the cloud provider name for digitalocean
	ProviderName = "digitalocean"

	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cloud.digitalocean.com/gpu-node"
)

// digitaloceanCloudProvider implements CloudProvider interface.
type digitaloceanCloudProvider struct {
	manager *DigitalOceanManager

	// nodeGroups contains the current set of node groups
	nodeGroups map[string]*NodeGroup

	// droplets contains a mapping of a node to a node groupd ID. Use the
	// nodeGroups map to obtain the actual node group
	droplets map[string]string
}

// Name returns name of the cloud provider.
func (d *digitaloceanCloudProvider) Name() string {
	return "digitalocean"
}

// NodeGroups returns all node groups configured for this cloud provider.
func (d *digitaloceanCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, 0, len(d.nodeGroups))
	for _, ng := range d.nodeGroups {
		nodeGroups = append(nodeGroups, ng)
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (d *digitaloceanCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	nodeGroupID, ok := d.droplets[node.Spec.ProviderID]
	if !ok {
		return nil, fmt.Errorf("node with id %q does not exist", node.Spec.ProviderID)
	}

	nodeGroup, ok := d.nodeGroups[nodeGroupID]
	if !ok {
		return nil, fmt.Errorf("node group with id %q does not exist", nodeGroupID)
	}

	return nodeGroup, nil
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (d *digitaloceanCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (d *digitaloceanCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (d *digitaloceanCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (d *digitaloceanCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GPULabel returns the label added to nodes with GPU resource.
func (d *digitaloceanCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (d *digitaloceanCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (d *digitaloceanCloudProvider) Cleanup() error {
	return cloudprovider.ErrNotImplemented
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups
// returned by NodeGroups() can change as a result of
// CloudProvider.Refresh().
func (d *digitaloceanCloudProvider) Refresh() error {
	return cloudprovider.ErrNotImplemented
}

// BuildDigitalOcean builds DigitalOcean cloud provider, manager etc.
func BuildDigitalOcean(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := NewDigitalOceanManager(config)
	if err != nil {
		klog.Fatalf("Failed to create DigitalOcean manager: %v", err)
	}

	provider, err := NewDigitalOceanCloudProvider(manager)
	if err != nil {
		klog.Fatalf("Failed to create DigitalOcean cloud provider: %v", err)
	}

	return provider
}

func NewDigitalOceanCloudProvider(manager *DigitalOceanManager) (*digitaloceanCloudProvider, error) {
	return &digitaloceanCloudProvider{
		manager: manager,
	}, nil
}
