package cloudresources

import (
	"context"
	"fmt"

	"github.com/nstogner/gcp-ilb-redirect-controller/cloud"

	compute "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type UrlMaps struct {
	Client   *cloud.Client
	resource *compute.UrlMap
}

func (c *UrlMaps) Create(ctx context.Context, info Info) error {
	res := &compute.UrlMap{
		Name: &info.GeneratedName,
		DefaultService: strptr(fmt.Sprintf("/compute/v1/projects/%s/regions/%s/backendServices/%s",
			info.Project, info.Region, info.GeneratedName)),
		HostRules: []*compute.HostRule{
			{
				Hosts:       []string{"*"},
				PathMatcher: strptr("all"),
			},
		},
		PathMatchers: []*compute.PathMatcher{
			{
				Name: strptr("all"),
				DefaultUrlRedirect: &compute.HttpRedirectAction{
					//HostRedirect:         strptr(info.IP + ":443"),
					HostRedirect:         strptr(info.Hostname),
					PathRedirect:         strptr("/"),
					RedirectResponseCode: strptr("PERMANENT_REDIRECT"),
					HttpsRedirect:        boolptr(true),
					StripQuery:           boolptr(false),
				},
			},
		},
	}

	_, err := c.Client.RegionUrlMaps.Insert(ctx, &compute.InsertRegionUrlMapRequest{
		UrlMapResource: res,
		Project:        info.Project,
		Region:         info.Region,
	})

	return err
}

func (c *UrlMaps) Delete(ctx context.Context, info Info) error {
	_, err := c.Client.RegionUrlMaps.Delete(ctx, &compute.DeleteRegionUrlMapRequest{
		UrlMap:  info.GeneratedName,
		Region:  info.Region,
		Project: info.Project,
	})
	return err
}

func (c *UrlMaps) Get(ctx context.Context, info Info) (err error) {
	c.resource, err = c.Client.RegionUrlMaps.Get(ctx, &compute.GetRegionUrlMapRequest{
		UrlMap:  info.GeneratedName,
		Region:  info.Region,
		Project: info.Project,
	})

	return
}
