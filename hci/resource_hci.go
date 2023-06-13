package hci

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/api"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

// GetHciResourceMap return the available Resource map
func GetHciResourceMap() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"hci_environment":          resourceHciEnvironment(),
		"hci_instance":             resourceHciInstance(),
		"hci_baremetal":            resourceHciBaremetal(),
		"hci_load_balancer_rule":   resourceHciLoadBalancerRule(),
		"hci_network":              resourceHciNetwork(),
		"hci_network_acl":          resourceHciNetworkACL(),
		"hci_network_acl_rule":     resourceHciNetworkACLRule(),
		"hci_port_forwarding_rule": resourceHciPortForwardingRule(),
		"hci_public_ip":            resourceHciPublicIP(),
		"hci_ssh_key":              resourceHciSSHKey(),
		"hci_static_nat":           resourceHciStaticNAT(),
		"hci_volume":               resourceHciVolume(),
		"hci_vpc":                  resourceHciVpc(),
		"hci_vpn":                  resourceHciVpn(),
		"hci_vpn_user":             resourceHciVpnUser(),
	}
}

func setValueOrID(d *schema.ResourceData, key string, value string, id string) error {
	if isID(d.Get(key).(string)) {
		return d.Set(key, id)
	}
	return d.Set(key, value)
}

func isID(id string) bool {
	re := regexp.MustCompile(`^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)
	return re.MatchString(id)
}

// Provides a common, simple way to deal with 404s.
func handleNotFoundError(entity string, deleted bool, err error, d *schema.ResourceData) error {
	if hciError, ok := err.(api.HciErrorResponse); ok {
		if hciError.StatusCode == 404 {
			d.SetId("")
			if deleted {
				log.Printf("%s (id=%s) not found", entity, d.Id())
				return nil
			}
			return fmt.Errorf("%s (id=%s) not found", entity, d.Id())
		}
	}
	return err
}

// Deals with all of the casting done to get a hci.Resources.
func getResources(d *schema.ResourceData, meta interface{}) hci.Resources {
	client := meta.(*hc.HciClient)
	_resources, _ := client.GetResources(d.Get("service_code").(string), d.Get("environment_name").(string))
	return _resources.(hci.Resources)
}

// Deals with all of the casting done to get a hci.Resources.
func getResourcesForEnvironmentID(client *hc.HciClient, environmentID string) (hci.Resources, error) {
	environment, err := client.Environments.Get(environmentID)
	if err != nil {
		return hci.Resources{}, err
	}
	resources, err := client.GetResources(environment.ServiceConnection.ServiceCode, environment.Name)
	if err != nil {
		return hci.Resources{}, err
	}
	return resources.(hci.Resources), nil
}
