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
type digitaloceanCloudProvider struct{}

func (d *digitaloceanCloudProvider) Name() string {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) NodeGroupForNode(*apiv1.Node) (cloudprovider.NodeGroup, error) {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) GPULabel() string {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) Cleanup() error {
	panic("not implemented")
}

func (d *digitaloceanCloudProvider) Refresh() error {
	panic("not implemented")
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

	// TODO(arslan): fill it
	return &digitaloceanCloudProvider{}
}
