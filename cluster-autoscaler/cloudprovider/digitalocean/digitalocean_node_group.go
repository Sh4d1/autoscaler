package digitalocean

import (
	"context"
	"fmt"
	"net/http"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean/godo"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	// These are internal DO values, not publicaly available and configurable
	// at this point.
	minNodePoolSize = 1
	maxNodePoolSize = 200
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id        string
	clusterID string
	client    nodeGroupClient
}

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return maxNodePoolSize
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return minNodePoolSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (n *NodeGroup) TargetSize() (int, error) {
	nodePool, err := n.getNodePool()
	if err != nil {
		return 0, err
	}

	return nodePool.Count, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	panic("not implemented")
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes([]*apiv1.Node) error {
	panic("not implemented")
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	panic("not implemented")
}

func (n *NodeGroup) Id() string {
	// Id returns an unique identifier of the node group.
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodePool, err := n.getNodePool()
	if err != nil {
		return nil, err
	}

	return toInstances(nodePool.Nodes), nil
}

// TemplateNodeInfo returns a schedulernodeinfo.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	panic("not implemented")
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (n *NodeGroup) Exist() bool {
	ctx := context.Background()
	_, resp, err := n.client.GetNodePool(ctx, n.clusterID, n.id)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false
		}

		klog.Errorf("couldn't obtain node pool information: %v", err)
		return false
	}

	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

func (n *NodeGroup) getNodePool() (*godo.KubernetesNodePool, error) {
	ctx := context.Background()
	nodePool, _, err := n.client.GetNodePool(ctx, n.clusterID, n.id)
	if err != nil {
		return nil, err
	}
	return nodePool, nil
}

func toInstances(nodes []*godo.KubernetesNode) []cloudprovider.Instance {
	instances := make([]cloudprovider.Instance, len(nodes))
	for i, nd := range nodes {
		instances[i] = toInstance(nd)
	}
	return instances
}

func toInstance(node *godo.KubernetesNode) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     node.ID,
		Status: toInstanceStatus(node.Status),
	}
}

func toInstanceStatus(nodeState *godo.KubernetesNodeStatus) *cloudprovider.InstanceStatus {
	if nodeState == nil {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch nodeState.State {
	case "provisioning":
		st.State = cloudprovider.InstanceCreating
	case "running":
		st.State = cloudprovider.InstanceRunning
	case "draining", "deleting":
		st.State = cloudprovider.InstanceDeleting
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-digitalocean",
			ErrorMessage: nodeState.Message,
		}
	}

	return st
}
