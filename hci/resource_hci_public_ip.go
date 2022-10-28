package hci

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciPublicIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciPublicIPCreate,
		Read:   resourceHciPublicIPRead,
		Delete: resourceHciPublicIPDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where the public IP should be created",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the VPC",
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHciPublicIPCreate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	vpcID := d.Get("vpc_id").(string)

	publicIPToCreate := hci.PublicIp{
		VpcId: vpcID,
	}
	newPublicIP, err := hciResources.PublicIps.Acquire(publicIPToCreate)
	if err != nil {
		return fmt.Errorf("Error acquiring the new public IP %s", err)
	}
	d.SetId(newPublicIP.Id)
	return resourceHciPublicIPRead(d, meta)
}

func resourceHciPublicIPRead(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}

	publicIP, err := hciResources.PublicIps.Get(d.Id())

	if err != nil {
		return handleNotFoundError("Public IP", false, err, d)
	}

	if err := d.Set("vpc_id", publicIP.VpcId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("ip_address", publicIP.IpAddress); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func resourceHciPublicIPDelete(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}

	if _, err := hciResources.PublicIps.Release(d.Id()); err != nil {
		return handleNotFoundError("Public IP", true, err, d)
	}

	return nil
}
