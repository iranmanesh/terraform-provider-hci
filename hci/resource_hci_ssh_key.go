package hci

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciSSHKey() *schema.Resource {
	return &schema.Resource{
		Create: createSSHKey,
		Read:   readSSHKey,
		Delete: deleteSSHKey,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where the SSH key should be created",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the SSH Key",
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func createSSHKey(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	name := d.Get("name").(string)
	publicKey := d.Get("public_key").(string)

	sk := hci.SSHKey{
		Name:      name,
		PublicKey: publicKey,
	}
	newSk, err := hciResources.SSHKeys.Create(sk)
	if err != nil {
		return fmt.Errorf("Error creating new SSH key %s", err)
	}
	d.SetId(newSk.ID)
	return readSSHKey(d, meta)
}

func readSSHKey(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}

	sk, err := hciResources.SSHKeys.Get(d.Id())

	if err != nil {
		return handleNotFoundError("SSH key", false, err, d)
	}

	if err := d.Set("name", sk.Name); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func deleteSSHKey(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}

	if _, err := hciResources.SSHKeys.Delete(d.Id()); err != nil {
		return handleNotFoundError("SSH key", true, err, d)
	}

	return nil
}
