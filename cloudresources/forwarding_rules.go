package cloudresources

import (
	"context"
	"fmt"

	"github.com/nstogner/gcp-ilb-redirect-controller/cloud"

	compute "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type ForwardingRules struct {
	Client   *cloud.Client
	resource *compute.ForwardingRule
}

func (c *ForwardingRules) Create(ctx context.Context, info Info) error {
	res := &compute.ForwardingRule{
		Name:                &info.GeneratedName,
		LoadBalancingScheme: strptr("INTERNAL_MANAGED"),
		Target: strptr(fmt.Sprintf("/compute/v1/projects/%s/regions/%s/targetHttpProxies/%s",
			info.Project, info.Region, info.GeneratedName)),
		Network:    strptr(fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/default", info.Project)),
		IPAddress:  strptr(info.IP),
		IPProtocol: strptr("TCP"),
		PortRange:  strptr("80"),
	}

	_, err := c.Client.ForwardingRules.Insert(ctx, &compute.InsertForwardingRuleRequest{
		ForwardingRuleResource: res,
		Project:                info.Project,
		Region:                 info.Region,
	})

	return err
}

func (c *ForwardingRules) Get(ctx context.Context, info Info) (err error) {
	c.resource, err = c.Client.ForwardingRules.Get(ctx, &compute.GetForwardingRuleRequest{
		ForwardingRule: info.GeneratedName,
		Region:         info.Region,
		Project:        info.Project,
	})

	return
}

func (c *ForwardingRules) Delete(ctx context.Context, info Info) error {
	_, err := c.Client.ForwardingRules.Delete(ctx, &compute.DeleteForwardingRuleRequest{
		ForwardingRule: info.GeneratedName,
		Region:         info.Region,
		Project:        info.Project,
	})
	return err
}
