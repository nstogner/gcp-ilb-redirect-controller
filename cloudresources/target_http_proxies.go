package cloudresources

import (
	"context"
	"fmt"

	"github.com/nstogner/gcp-ilb-redirect-controller/cloud"

	compute "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type TargetHttpProxies struct {
	Client   *cloud.Client
	resource *compute.TargetHttpProxy
}

func (c *TargetHttpProxies) Create(ctx context.Context, info Info) error {
	res := &compute.TargetHttpProxy{
		Name: &info.GeneratedName,
		UrlMap: strptr(fmt.Sprintf("/compute/v1/projects/%s/regions/%s/urlMaps/%s",
			info.Project, info.Region, info.GeneratedName)),
	}

	_, err := c.Client.RegionTargetHttpProxies.Insert(ctx, &compute.InsertRegionTargetHttpProxyRequest{
		TargetHttpProxyResource: res,
		Project:                 info.Project,
		Region:                  info.Region,
	})

	return err
}

func (c *TargetHttpProxies) Get(ctx context.Context, info Info) (err error) {
	c.resource, err = c.Client.RegionTargetHttpProxies.Get(ctx, &compute.GetRegionTargetHttpProxyRequest{
		TargetHttpProxy: info.GeneratedName,
		Region:          info.Region,
		Project:         info.Project,
	})

	return
}

func (c *TargetHttpProxies) Delete(ctx context.Context, info Info) error {
	_, err := c.Client.RegionTargetHttpProxies.Delete(ctx, &compute.DeleteRegionTargetHttpProxyRequest{
		TargetHttpProxy: info.GeneratedName,
		Region:          info.Region,
		Project:         info.Project,
	})
	return err
}
