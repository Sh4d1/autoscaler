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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"golang.org/x/oauth2"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean/godo"
)

type nodeGroupClient interface {
	GetNodePool(ctx context.Context, clusterID, poolID string) (*godo.KubernetesNodePool, *godo.Response, error)
	ListNodePools(ctx context.Context, clusterID string, opts *godo.ListOptions) ([]*godo.KubernetesNodePool, *godo.Response, error)
}

// DigitalOceanManager handles DigitalOcean communication and data caching of
// node groups (node pools in DOKS)
type DigitalOceanManager struct {
	client nodeGroupClient

	clusterID string

	// nodeGroups contains the current set of node groups
	nodeGroups map[string]*NodeGroup

	// droplets contains a mapping of a node to a node group ID. Use the
	// nodeGroups map to obtain the actual node group
	droplets map[string]string
}

// Config is the configuration of the DigitalOcean cloud provider
type Config struct {
	// clusterID is the id associated with the cluster where DigitalOcean
	// Cluster Autoscaler is running.
	clusterID string

	// token is the User's Access Token associated with the cluster where
	// DigitalOcean Cluster Autoscaler is running.
	token string

	// url points to DigitalOcean API. If empty, defaults to
	// https://api.digitalocean.com/
	url string

	// version defines the version of the DigitalOcean cluster autoscaler
	version string
}

func NewDigitalOceanManager(configReader io.Reader) (*DigitalOceanManager, error) {
	cfg := &Config{}
	if configReader != nil {
		body, err := ioutil.ReadAll(configReader)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.token == "" {
		return nil, errors.New("access token  is not provided")
	}
	if cfg.clusterID == "" {
		return nil, errors.New("cluster ID is not provided")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: cfg.token,
	})
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)

	opts := []godo.ClientOpt{}
	if cfg.url != "" {
		opts = append(opts, godo.SetBaseURL(cfg.url))
	}

	version := "dev"
	if cfg.version != "" {
		version = cfg.version
	}
	opts = append(opts, godo.SetUserAgent("cluster-autoscaler-digitalocean/"+version))

	doClient, err := godo.New(oauthClient, opts...)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize DigitalOcean client: %s", err)
	}

	d := &DigitalOceanManager{
		client:     doClient.Kubernetes,
		clusterID:  cfg.clusterID,
		nodeGroups: make(map[string]*NodeGroup, 0),
		droplets:   make(map[string]string, 0),
	}

	// initialize the node groups
	if err := d.Refresh(); err != nil {
		return nil, err
	}

	return d, nil
}

// Refresh refreshes the cache holding the nodegroups
func (d *DigitalOceanManager) Refresh() error {
	ctx := context.Background()
	nodePools, _, err := d.client.ListNodePools(ctx, d.clusterID, nil)
	if err != nil {
		return err
	}

	for _, np := range nodePools {
		// NOTE(arslan): do not include the size or nodes in this struct as
		// those are dynamic and can change. Those will be handled
		// dynamically within the NodeGroup type
		d.nodeGroups[np.ID] = &NodeGroup{
			id:        np.ID,
			clusterID: d.clusterID,
			client:    d.client,
		}
	}

	return nil
}
