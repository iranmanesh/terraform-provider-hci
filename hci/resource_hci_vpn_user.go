package hci

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/api"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciVpnUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciVpnUserCreate,
		Read:   resourceHciVpnUserRead,
		Delete: resourceHciVpnUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the environment where the vpn should be created",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Username of the VPN user",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Password of the VPN user",
			},
		},
	}
}

func resourceHciVpnUserCreate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))
	if rerr != nil {
		return rerr
	}

	remoteAccessVpnUser := hci.RemoteAccessVpnUser{
		Username: d.Get("username").(string),
		Password: d.Get("password").(string),
	}
	_, err := hciResources.RemoteAccessVpnUser.Create(remoteAccessVpnUser)
	if err != nil {
		return fmt.Errorf("Error adding VPN user: %s", err)
	}

	// TODO: When the CMC API actually returns the ID of the created user, use it.
	// Currently there is no way to do a 'Get' based on the username, and we don't have the ID, so
	// we have to list all users and then loop through to match the username in order to find the ID.
	vpnUsers, err := hciResources.RemoteAccessVpnUser.List()
	if err != nil {
		return fmt.Errorf("Error getting the created VPN user ID: %s", err)
	}
	var userID string
	for _, user := range vpnUsers {
		if user.Username == d.Get("username").(string) {
			userID = user.Id
			break
		}
	}
	if userID != "" {
		d.SetId(userID)
	} else {
		return fmt.Errorf("Error finding the created VPN user ID: %s", err)
	}
	return resourceHciVpnUserRead(d, meta)
}

func resourceHciVpnUserRead(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))
	if rerr != nil {
		return rerr
	}

	// Get the user based on the ID
	vpnUser, err := hciResources.RemoteAccessVpnUser.Get(d.Id())
	if err != nil {
		d.SetId("")
		// If we return an error instead of nil, then if a VPN user is removed via the web UI
		// it will break the ability for terraform to plan or apply any changes, so terraform
		// will be in a broken state which can not be recovered from.
		return nil
	}

	if err := d.Set("username", vpnUser.Username); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}
	return nil
}

func resourceHciVpnUserDelete(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))
	if rerr != nil {
		return rerr
	}
	remoteAccessVpnUser := hci.RemoteAccessVpnUser{
		Id:       d.Id(),
		Username: d.Get("username").(string),
	}
	if _, err := hciResources.RemoteAccessVpnUser.Delete(remoteAccessVpnUser); err != nil {
		if hciError, ok := err.(api.HciErrorResponse); ok {
			if hciError.StatusCode == 404 {
				log.Printf("VPN User with id=%s no longer exists", d.Id())
				d.SetId("")
				return nil
			}
			return handleNotFoundError("VPN User Delete", true, err, d)
		}
		return handleNotFoundError("VPN User Delete", true, err, d)
	}
	return nil
}
