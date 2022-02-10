package cloudresources

import (
	"context"

	"github.com/nstogner/gcp-ilb-redirect-controller/cloud"

	compute "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type BackendServices struct {
	Client   *cloud.Client
	resource *compute.BackendService
}

func (c *BackendServices) Create(ctx context.Context, info Info) error {
	res := &compute.BackendService{
		Name:                &info.GeneratedName,
		LoadBalancingScheme: strptr("INTERNAL_MANAGED"),
		Protocol:            strptr("HTTP"),
	}

	_, err := c.Client.RegionBackendServices.Insert(ctx, &compute.InsertRegionBackendServiceRequest{
		BackendServiceResource: res,
		Project:                info.Project,
		Region:                 info.Region,
	})

	return err
}

func (c *BackendServices) Get(ctx context.Context, info Info) (err error) {
	c.resource, err = c.Client.RegionBackendServices.Get(ctx, &compute.GetRegionBackendServiceRequest{
		BackendService: info.GeneratedName,
		Region:         info.Region,
		Project:        info.Project,
	})

	return
}

func (c *BackendServices) Delete(ctx context.Context, info Info) error {
	_, err := c.Client.RegionBackendServices.Delete(ctx, &compute.DeleteRegionBackendServiceRequest{
		BackendService: info.GeneratedName,
		Region:         info.Region,
		Project:        info.Project,
	})
	return err
}
