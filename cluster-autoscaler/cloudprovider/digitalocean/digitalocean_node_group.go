package digitalocean

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// NodeGroup implements cloudprovider.NodeGroup interface
type NodeGroup struct {
	id string
}

func (n *NodeGroup) MaxSize() int {
	panic("not implemented")
}

func (n *NodeGroup) MinSize() int {
	panic("not implemented")
}

func (n *NodeGroup) TargetSize() (int, error) {
	panic("not implemented")
}

func (n *NodeGroup) IncreaseSize(delta int) error {
	panic("not implemented")
}

func (n *NodeGroup) DeleteNodes([]*apiv1.Node) error {
	panic("not implemented")
}

func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	panic("not implemented")
}

func (n *NodeGroup) Id() string {
	return n.id
}

func (n *NodeGroup) Debug() string {
	panic("not implemented")
}

func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	panic("not implemented")
}

func (n *NodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	panic("not implemented")
}

func (n *NodeGroup) Exist() bool {
	return true
}

func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

func (n *NodeGroup) Autoprovisioned() bool {
	panic("not implemented")
}
