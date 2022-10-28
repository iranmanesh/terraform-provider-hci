package hci

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

func resourceHciVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciVolumeCreate,
		Read:   resourceHciVolumeRead,
		Update: resourceHciVolumeUpdate,
		Delete: resourceHciVolumeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of environment where the volume should be created",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the volume to be created",
			},
			"disk_offering": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID or name of the disk offering of the new volume",
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"size_in_gb": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The size of the volume in gigabytes",
			},
			"iops": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The number of iops of the volume",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the instance to which the volume will be attached",
			},
		},
	}
}

func resourceHciVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	diskOffering, err := retrieveDiskOffering(&hciResources, d.Get("disk_offering").(string))
	if err != nil {
		return err
	}
	volumeToCreate := hci.Volume{
		Name:           d.Get("name").(string),
		DiskOfferingId: diskOffering.Id,
	}

	if val, ok := d.GetOk("size_in_gb"); ok {
		if !diskOffering.CustomSize {
			return fmt.Errorf("Disk offering %s doesn't allow custom size", diskOffering.Id)
		}
		volumeToCreate.GbSize = val.(int)
	}

	if val, ok := d.GetOk("iops"); ok {
		if !diskOffering.CustomIops {
			return fmt.Errorf("Disk offering %s doesn't allow custom IOPS", diskOffering.Id)
		}
		volumeToCreate.Iops = val.(int)
	}

	if zone, ok := d.GetOk("zone"); ok {
		if isID(zone.(string)) {
			volumeToCreate.ZoneId = zone.(string)
		} else {
			volumeToCreate.ZoneId, err = retrieveZoneID(&hciResources, zone.(string))
			if err != nil {
				return err
			}
		}
	}

	if instanceID, ok := d.GetOk("instance_id"); ok {
		volumeToCreate.InstanceId = instanceID.(string)
	}

	newVolume, err := hciResources.Volumes.Create(volumeToCreate)
	if err != nil {
		return err
	}
	d.SetId(newVolume.Id)
	return resourceHciVolumeRead(d, meta)
}

func resourceHciVolumeRead(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	volume, err := hciResources.Volumes.Get(d.Id())
	if err != nil {
		return handleNotFoundError("Volume", false, err, d)
	}

	if err := d.Set("name", volume.Name); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := setValueOrID(d, "disk_offering", strings.ToLower(volume.DiskOfferingName), volume.DiskOfferingId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("size_in_gb", volume.GbSize); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("iops", volume.Iops); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set("instance_id", volume.InstanceId); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func resourceHciVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	d.Partial(true)
	curVolume, err := hciResources.Volumes.Get(d.Id())
	if err != nil {
		return err
	}
	if d.HasChange("instance_id") {
		oldInstanceID, newInstanceID := d.GetChange("instance_id")
		volume := &hci.Volume{
			Id: d.Id(),
		}
		if oldInstanceID != "" && curVolume.InstanceId != "" {
			err := hciResources.Volumes.DetachFromInstance(volume)
			if err != nil {
				return err
			}
		}
		if newInstanceID != "" {
			err := hciResources.Volumes.AttachToInstance(volume, newInstanceID.(string))
			if err != nil {
				return err
			}
		}
	}
	if d.HasChange("size_in_gb") || d.HasChange("iops") {
		volumeToResize := hci.Volume{
			Id: d.Id(),
		}
		if val, ok := d.GetOk("size_in_gb"); ok {
			volumeToResize.GbSize = val.(int)
			if curVolume.GbSize > volumeToResize.GbSize {
				return fmt.Errorf("Cannot reduce size of a volume")
			}
		}
		if val, ok := d.GetOk("iops"); ok {
			volumeToResize.Iops = val.(int)
		}
		_ = hciResources.Volumes.Resize(&volumeToResize)
	}
	d.Partial(false)
	return resourceHciVolumeRead(d, meta)
}

func resourceHciVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	hciResources, rerr := getResourcesForEnvironmentID(meta.(*hc.HciClient), d.Get("environment_id").(string))

	if rerr != nil {
		return rerr
	}
	if instanceID, ok := d.GetOk("instance_id"); ok && instanceID != "" {
		volume := &hci.Volume{
			Id: d.Id(),
		}
		err := hciResources.Volumes.DetachFromInstance(volume)
		if err != nil {
			return err
		}
	}
	if err := hciResources.Volumes.Delete(d.Id()); err != nil {
		return handleNotFoundError("Volume", true, err, d)
	}
	return nil
}

func retrieveZoneID(hciResources *hci.Resources, zoneName string) (zoneID string, nerr error) {
	zones, err := hciResources.Zones.List()
	if err != nil {
		return "", err
	}
	for _, zone := range zones {
		if strings.EqualFold(zone.Name, zoneName) {
			return zone.Id, nil
		}
	}
	return "", fmt.Errorf("Zone with name %s could not be found", zoneName)
}

func retrieveDiskOffering(hciRes *hci.Resources, name string) (diskOffering *hci.DiskOffering, err error) {
	if isID(name) {
		return hciRes.DiskOfferings.Get(name)
	}
	offerings, err := hciRes.DiskOfferings.List()
	if err != nil {
		return nil, err
	}
	for _, offering := range offerings {
		if strings.EqualFold(offering.Name, name) {
			log.Printf("Found disk offering: %+v", offering)
			return &offering, nil
		}
	}
	return nil, fmt.Errorf("Disk offering with name %s not found", name)
}
