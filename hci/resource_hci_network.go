package hci

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/api"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciNetworkCreate,
		Read:   resourceHciNetworkRead,
		Update: resourceHciNetworkUpdate,
		Delete: resourceHciNetworkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where network should be created",
			},
			"organization_code": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Entry point of organization",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of network",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of network",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the VPC",
			},
			"network_offering": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `The network offering name or id (e.g. "Standard Network" or "Load Balanced Network")`,
			},
			"network_acl": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name or id of the network ACL",
			},
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHciNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	networkOfferingID, nerr := retrieveNetworkOfferingID(&hciResources, d.Get("network_offering").(string))
	if nerr != nil {
		return nerr
	}

	aclID, nerr := retrieveNetworkACLID(&hciResources, d.Get("network_acl").(string), d.Get("vpc_id").(string))
	if nerr != nil {
		return nerr
	}

	networkToCreate := hci.Network{
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		VpcId:             d.Get("vpc_id").(string),
		NetworkOfferingId: networkOfferingID,
		NetworkAclId:      aclID,
	}
	options := map[string]string{}
	if orgID, ok := d.GetOk("organization_code"); ok {
		options["org_id"] = orgID.(string)
	}
	newNetwork, err := hciResources.Networks.Create(networkToCreate, options)
	if err != nil {
		return fmt.Errorf("Error creating the new network %s: %s", networkToCreate.Name, err)
	}
	d.SetId(newNetwork.Id)
	return resourceHciNetworkRead(d, meta)
}

func resourceHciNetworkRead(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	network, err := hciResources.Networks.Get(d.Id())
	if err != nil {
		if hciError, ok := err.(api.HciErrorResponse); ok {
			if hciError.StatusCode == 404 {
				log.Printf("Network %s was not found", d.Get("name").(string))
				d.SetId("")
				return nil
			}
		}
		return err
	}

	offering, offErr := hciResources.NetworkOfferings.Get(network.NetworkOfferingId)
	if offErr != nil {
		return handleNotFoundError("Network", false, offErr, d)
	}

	// Update the config
	if err := d.Set("name", network.Name); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("description", network.Description); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := setValueOrID(d, "network_offering", offering.Name, network.NetworkOfferingId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("vpc_id", network.VpcId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := setValueOrID(d, "network_acl", network.NetworkAclName, network.NetworkAclId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("cidr", network.Cidr); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func resourceHciNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	d.Partial(true)

	if d.HasChange("name") || d.HasChange("description") {
		newName := d.Get("name").(string)
		newDescription := d.Get("description").(string)
		_, err := hciResources.Networks.Update(d.Id(), hci.Network{Id: d.Id(), Name: newName, Description: newDescription})
		if err != nil {
			return err
		}
	}

	if d.HasChange("network_acl") {
		aclID, err := retrieveNetworkACLID(&hciResources, d.Get("network_acl").(string), d.Get("vpc_id").(string))
		if err != nil {
			return err
		}
		_, aclErr := hciResources.Networks.ChangeAcl(d.Id(), aclID)
		if aclErr != nil {
			return aclErr
		}
	}

	d.Partial(false)

	return nil
}

func resourceHciNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	if _, err := hciResources.Networks.Delete(d.Id()); err != nil {
		return handleNotFoundError("Network", true, err, d)
	}

	return nil
}

func retrieveNetworkOfferingID(hciRes *hci.Resources, name string) (id string, err error) {
	if isID(name) {
		return name, nil
	}
	offerings, err := hciRes.NetworkOfferings.List()
	if err != nil {
		return "", err
	}
	for _, offering := range offerings {
		if strings.EqualFold(offering.Name, name) {
			log.Printf("Found network offering: %+v", offering)
			return offering.Id, nil
		}
	}
	return "", fmt.Errorf("Network offering with name %s not found", name)
}

func retrieveNetworkACLID(hciRes *hci.Resources, name, vpcID string) (id string, err error) {
	if isID(name) {
		return name, nil
	}
	acls, err := hciRes.NetworkAcls.ListByVpcId(vpcID)
	if err != nil {
		return "", err
	}
	for _, acl := range acls {
		if strings.EqualFold(acl.Name, name) {
			return acl.Id, nil
		}
	}
	return "", fmt.Errorf("Network ACL with name %s not found", name)
}
