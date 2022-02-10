package cloud

import (
	"context"
	"fmt"
	"net/http"

	compute "cloud.google.com/go/compute/apiv1"
	"google.golang.org/api/googleapi"
)

type Client struct {
	RegionBackendServices   *compute.RegionBackendServicesClient
	RegionUrlMaps           *compute.RegionUrlMapsClient
	RegionTargetHttpProxies *compute.RegionTargetHttpProxiesClient
	ForwardingRules         *compute.ForwardingRulesClient
}

func NewClient(ctx context.Context) (*Client, error) {
	backendServices, err := compute.NewRegionBackendServicesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewBackendServicesRESTClient: %w", err)
	}

	urlMaps, err := compute.NewRegionUrlMapsRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewRegionUrlMapsRESTClient: %w", err)
	}

	targetHttpProxies, err := compute.NewRegionTargetHttpProxiesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewRegionTargetHttpProxiesRESTClient: %w", err)
	}

	forwardingRules, err := compute.NewForwardingRulesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewForwardingRulesRESTClient: %w", err)
	}

	return &Client{
		RegionBackendServices:   backendServices,
		RegionUrlMaps:           urlMaps,
		RegionTargetHttpProxies: targetHttpProxies,
		ForwardingRules:         forwardingRules,
	}, nil
}

func (c *Client) Close() {
	c.RegionBackendServices.Close()
	c.RegionUrlMaps.Close()
	c.RegionTargetHttpProxies.Close()
	c.ForwardingRules.Close()
}

func IsNotFound(err error) bool {
	if gerr, ok := err.(*googleapi.Error); ok {
		return gerr.Code == http.StatusNotFound
	}

	return false
}

func IsNotReady(err error) bool {
	if gerr, ok := err.(*googleapi.Error); ok {
		for _, item := range gerr.Errors {
			if item.Reason == "resourceNotReady" {
				return true
			}
		}
	}

	return false
}
