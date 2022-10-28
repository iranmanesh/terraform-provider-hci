package hci

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciPortForwardingRule() *schema.Resource {
	return &schema.Resource{
		Create: createPortForwardingRule,
		Read:   readPortForwardingRule,
		Delete: deletePortForwardingRule,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where port forwarding rule should be created",
			},
			"public_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The public IP to which these rules should be applied",
			},
			"private_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the private IP to bind to",
			},
			"protocol": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The protocol that this rule should use (eg. TCP, UDP)",
			},
			"private_port_start": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The start of the private port range for this rule",
			},
			"private_port_end": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "The end of the private port range for this rule",
			},
			"public_port_start": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The start of the public port range for this rule",
			},
			"public_port_end": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "The end of the public port range for this rule",
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func createPortForwardingRule(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	pfr := hci.PortForwardingRule{
		PublicIpId:       d.Get("public_ip_id").(string),
		Protocol:         d.Get("protocol").(string),
		PublicPortStart:  d.Get("public_port_start").(string),
		PrivateIpId:      d.Get("private_ip_id").(string),
		PrivatePortStart: d.Get("private_port_start").(string),
	}

	if _, ok := d.GetOk("public_port_end"); ok {
		pfr.PublicPortEnd = d.Get("public_port_end").(string)
	}

	if _, ok := d.GetOk("private_port_end"); ok {
		pfr.PrivatePortEnd = d.Get("private_port_end").(string)
	}

	newPfr, err := hciResources.PortForwardingRules.Create(pfr)
	if err != nil {
		return err
	}

	d.SetId(newPfr.Id)
	return readPortForwardingRule(d, meta)
}

func readPortForwardingRule(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	pfr, err := hciResources.PortForwardingRules.Get(d.Id())
	if err != nil {
		return handleNotFoundError("Port forwarding rule", false, err, d)
	}

	if err := d.Set("public_ip_id", pfr.PublicIpId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("private_ip_id", pfr.PrivateIpId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("instance_id", pfr.InstanceId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("protocol", pfr.Protocol); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("public_port_start", pfr.PublicPortStart); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("public_port_end", pfr.PublicPortEnd); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("private_port_start", pfr.PrivatePortStart); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("private_port_end", pfr.PrivatePortEnd); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("private_ip", pfr.PrivateIp); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("public_ip", pfr.PublicIp); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func deletePortForwardingRule(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	if _, err := hciResources.PortForwardingRules.Delete(d.Id()); err != nil {
		return handleNotFoundError("Port forwarding rule", true, err, d)
	}
	return nil
}
