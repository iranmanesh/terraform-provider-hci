package hci

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciStaticNAT() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciStaticNATCreate,
		Read:   resourceHciStaticNATRead,
		Delete: resourceHciStaticNATDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where static NAT should be enabled",
			},
			"public_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The public IP to enable static NAT on",
			},
			"private_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The private IP to enable static NAT on",
			},
		},
	}
}

func resourceHciStaticNATCreate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	staticNATPublicIP := hci.PublicIp{
		Id:          d.Get("public_ip_id").(string),
		PrivateIpId: d.Get("private_ip_id").(string),
	}
	_, err := hciResources.PublicIps.EnableStaticNat(staticNATPublicIP)
	if err != nil {
		return fmt.Errorf("Error enabling static NAT: %s", err)
	}
	d.SetId(staticNATPublicIP.Id)
	return resourceHciStaticNATRead(d, meta)
}

func resourceHciStaticNATRead(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	publicIP, err := hciResources.PublicIps.Get(d.Id())
	if err != nil {
		return handleNotFoundError("Static NAT", false, err, d)
	}
	if publicIP.PrivateIpId == "" {
		// If the private IP ID is missing, it means the public IP no longer has static NAT
		// enabled and so this entity is "missing" (at least as far as terraform is concerned).
		d.SetId("")
		return nil
	}
	if err := d.Set("private_ip_id", publicIP.PrivateIpId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}
	return nil
}

func resourceHciStaticNATDelete(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	_, err := hciResources.PublicIps.DisableStaticNat(d.Id())
	return handleNotFoundError("Static NAT", true, err, d)
}
