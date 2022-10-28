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

func resourceHciVpc() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciVpcCreate,
		Read:   resourceHciVpcRead,
		Update: resourceHciVpcUpdate,
		Delete: resourceHciVpcDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where VPC should be created",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of VPC",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of VPC",
			},
			"vpc_offering": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name or id of the VPC offering",
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"network_domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "A custom DNS suffix at the level of a network",
			},
			"zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "Zone ID or name where the VPC is created",
			},
		},
	}
}

func resourceHciVpcCreate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	vpcOfferingID, cerr := retrieveVpcOfferingID(&hciResources, d.Get("vpc_offering").(string))

	if cerr != nil {
		return cerr
	}

	vpcToCreate := hci.Vpc{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		VpcOfferingId: vpcOfferingID,
	}

	if networkDomain, ok := d.GetOk("network_domain"); ok {
		vpcToCreate.NetworkDomain = networkDomain.(string)
	}

	if zone, ok := d.GetOk("zone"); ok {
		if isID(zone.(string)) {
			vpcToCreate.ZoneId = zone.(string)
		} else {
			var zErr error
			vpcToCreate.ZoneId, zErr = retrieveZoneID(&hciResources, zone.(string))
			if zErr != nil {
				return zErr
			}
		}
	}

	newVpc, err := hciResources.Vpcs.Create(vpcToCreate)
	if err != nil {
		return fmt.Errorf("Error creating the new VPC %s: %s", vpcToCreate.Name, err)
	}
	d.SetId(newVpc.Id)

	return resourceHciVpcRead(d, meta)
}

func resourceHciVpcRead(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	// Get the vpc details
	vpc, err := hciResources.Vpcs.Get(d.Id())
	if err != nil {
		return handleNotFoundError("VPC", false, err, d)
	}

	if err := setValueOrID(d, "zone", vpc.ZoneName, vpc.ZoneId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	vpcOffering, offErr := hciResources.VpcOfferings.Get(vpc.VpcOfferingId)
	if offErr != nil {
		if hciError, ok := offErr.(api.HciErrorResponse); ok {
			if hciError.StatusCode == 404 {
				log.Printf("VPC offering id=%s does no longer exist", vpc.VpcOfferingId)
				d.SetId("")
				return nil
			}
		}
		return offErr
	}

	// Update the config
	if err := d.Set("name", vpc.Name); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("description", vpc.Description); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := setValueOrID(d, "vpc_offering", strings.ToLower(vpcOffering.Name), vpc.VpcOfferingId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("network_domain", vpc.NetworkDomain); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func resourceHciVpcUpdate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	if d.HasChange("name") || d.HasChange("description") {
		newName := d.Get("name").(string)
		newDescription := d.Get("description").(string)
		log.Printf("[DEBUG] Details have changed updating VPC.....")
		_, err := hciResources.Vpcs.Update(hci.Vpc{Id: d.Id(), Name: newName, Description: newDescription})
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceHciVpcDelete(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	fmt.Printf("[INFO] Destroying VPC: %s\n", d.Get("name").(string))
	if _, err := hciResources.Vpcs.Destroy(d.Id()); err != nil {
		return handleNotFoundError("VPC", true, err, d)
	}

	return nil
}

func retrieveVpcOfferingID(hciRes *hci.Resources, name string) (id string, err error) {
	if isID(name) {
		return name, nil
	}

	vpcOfferings, err := hciRes.VpcOfferings.List()
	if err != nil {
		return "", err
	}
	for _, offering := range vpcOfferings {
		if strings.EqualFold(offering.Name, name) {
			log.Printf("Found vpc offering: %+v", offering)
			return offering.Id, nil
		}
	}

	return "", fmt.Errorf("VPC offering with name %s not found", name)
}
