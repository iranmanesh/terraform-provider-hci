package hci

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciBaremetal() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciBaremetalCreate,
		Read:   resourceHciBaremetalRead,
		Update: resourceHciBaremetalUpdate,
		Delete: resourceHciBaremetalDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where baremetal should be created",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of baremetal",
			},
			"template": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name or id of the template to use for this baremetal",
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"compute_offering": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name or id of the compute offering to use for this baremetal",
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the network into which the new baremetal will be created",
			},
			"hypervisor": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of hypervisor",
			},
			"ssh_key_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SSH key name to attach to the new baremetal. Note: Cannot be used with public key.",
			},
			"public_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Public key to attach to the new baremetal. Note: Cannot be used with SSH key name.",
			},
			"user_data": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional data passed to the new baremetal during its initialization",
			},
			"private_ip_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The IPv4 address of the baremetal. Must be within the network's CIDR and not collide with existing baremetals.",
			},
			"dedicated_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Id of the dedicated group into which the new baremetal will be created",
			},
		},
	}
}

func resourceHciBaremetalCreate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}

	computeOfferingID, cerr := retrieveComputeOfferingID(&hciResources, d.Get("compute_offering").(string))

	if cerr != nil {
		return cerr
	}

	templateID, terr := retrieveTemplateID(&hciResources, d.Get("template").(string))

	if terr != nil {
		return terr
	}

	baremetalToCreate := hci.Baremetal{Name: d.Get("name").(string),
		ComputeOfferingId: computeOfferingID,
		TemplateId:        templateID,
		ImageId:           templateID, // TODO This bug must be fixed in CloudMC API
		NetworkId:         d.Get("network_id").(string),
	}

	if sshKeyname, ok := d.GetOk("ssh_key_name"); ok {
		baremetalToCreate.SSHKeyName = sshKeyname.(string)
	}
	if publicKey, ok := d.GetOk("public_key"); ok {
		baremetalToCreate.PublicKey = publicKey.(string)
	}
	if userData, ok := d.GetOk("user_data"); ok {
		baremetalToCreate.UserData = userData.(string)
	}
	if privateIP, ok := d.GetOk("private_ip"); ok {
		baremetalToCreate.IpAddress = privateIP.(string)
	}
	baremetalToCreate.Hypervisor = "Baremetal"

	computeOffering, cerr := hciResources.ComputeOfferings.Get(computeOfferingID)
	_ = computeOffering
	if cerr != nil {
		return cerr
	}

	if dedicatedGroupID, ok := d.GetOk("dedicated_group_id"); ok {
		baremetalToCreate.DedicatedGroupId = dedicatedGroupID.(string)
	}

	newBaremetal, err := hciResources.Baremetals.Create(baremetalToCreate)
	if err != nil {
		return fmt.Errorf("Error creating the new baremetal %s: %s", baremetalToCreate.Name, err)
	}

	d.SetId(newBaremetal.Id)
	d.SetConnInfo(map[string]string{
		"host":     newBaremetal.IpAddress,
		"user":     newBaremetal.Username,
		"password": newBaremetal.Password,
	})

	return resourceHciBaremetalRead(d, meta)
}

func resourceHciBaremetalRead(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	// Get the virtual machine details
	baremetal, err := hciResources.Baremetals.Get(d.Id())
	if err != nil {
		return handleNotFoundError("Baremetal", false, err, d)
	}
	// Update the config
	if err := d.Set("name", baremetal.Name); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := setValueOrID(d, "template", strings.ToLower(baremetal.TemplateName), baremetal.TemplateId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := setValueOrID(d, "compute_offering", strings.ToLower(baremetal.ComputeOfferingName), baremetal.ComputeOfferingId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("network_id", baremetal.NetworkId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("private_ip_id", baremetal.IpAddressId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("private_ip", baremetal.IpAddress); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	dID, dIDErr := getBMDedicatedGroupID(hciResources, baremetal)
	if dIDErr != nil {
		return dIDErr
	}

	if err := d.Set("dedicated_group_id", dID); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func resourceHciBaremetalUpdate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	d.Partial(true)

	if d.HasChange("ssh_key_name") {
		sshKeyName := d.Get("ssh_key_name").(string)
		log.Printf("[DEBUG] SSH key name has changed for %s, associating new SSH key...", sshKeyName)
		_, err := hciResources.Baremetals.AssociateSSHKey(d.Id(), sshKeyName)
		if err != nil {
			return err
		}
	}

	if d.HasChange("private_ip") {
		return fmt.Errorf("Cannot update the private IP of a baremetal")
	}

	d.Partial(false)

	return nil
}

func resourceHciBaremetalDelete(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	fmt.Printf("[INFO] Destroying baremetal: %s\n", d.Get("name").(string))
	if _, err := hciResources.Baremetals.Destroy(d.Id()); err != nil {
		return handleNotFoundError("Baremetal", true, err, d)
	}

	return nil
}

func getBMDedicatedGroupID(hciRes hci.Resources, baremetal *hci.Baremetal) (string, error) {
	dedicatedGroups, err := hciRes.AffinityGroups.ListWithOptions(map[string]string{
		"type": "ExplicitDedication",
	})
	if err != nil {
		return "", err
	}
	for _, dedicatedGroup := range dedicatedGroups {
		for _, affinityGroupID := range baremetal.AffinityGroupIds {
			if strings.EqualFold(dedicatedGroup.Id, affinityGroupID) {
				return dedicatedGroup.Id, nil
			}
		}
	}
	return "", nil
}
