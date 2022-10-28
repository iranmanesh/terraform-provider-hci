package hci

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciNetworkACL() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciNetworkACLCreate,
		Read:   resourceHciNetworkACLRead,
		Delete: resourceHciNetworkACLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where the network ACL should be created",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of network ACL",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Description of network ACL",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the VPC",
			},
		},
	}
}

func resourceHciNetworkACLCreate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}

	aclToCreate := hci.NetworkAcl{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		VpcId:       d.Get("vpc_id").(string),
	}
	newACL, err := hciResources.NetworkAcls.Create(aclToCreate)
	if err != nil {
		return fmt.Errorf("Error creating the new network ACL %s: %s", aclToCreate.Name, err)
	}
	d.SetId(newACL.Id)
	return resourceHciNetworkACLRead(d, meta)
}

func resourceHciNetworkACLRead(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	acl, aErr := hciResources.NetworkAcls.Get(d.Id())
	if aErr != nil {
		return handleNotFoundError("Network ACL", false, aErr, d)
	}

	// Update the config
	if err := d.Set("name", acl.Name); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("description", acl.Description); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("vpc_id", acl.VpcId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func resourceHciNetworkACLDelete(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	if _, err := hciResources.NetworkAcls.Delete(d.Id()); err != nil {
		return handleNotFoundError("Network ACL", true, err, d)
	}
	return nil
}
