package hci

import (
	"encoding/json"
	"strings"

	"github.com/hypertec-cloud/go-hci/api"
	"github.com/hypertec-cloud/go-hci/services"
)

const (
	BAREMETAL_STATE_RUNNING = "Running"
	BAREMETAL_STATE_STOPPED = "Stopped"
)

const (
	BAREMETAL_START_OPERATION                   = "start"
	BAREMETAL_STOP_OPERATION                    = "stop"
	BAREMETAL_REBOOT_OPERATION                  = "reboot"
	BAREMETAL_RECOVER_OPERATION                 = "recover"
	BAREMETAL_PURGE_OPERATION                   = "releaseBareMetal"
	BAREMETAL_CHANGE_NETWORK_OFFERING_OPERATION = "changeNetwork"
	BAREMETAL_ASSOCIATE_SSH_KEY_OPERATION       = "associateSSHKey"
)

type Baremetal struct {
	Id                       string        `json:"id,omitempty"`
	Name                     string        `json:"name,omitempty"`
	State                    string        `json:"state,omitempty"`
	TemplateId               string        `json:"templateId,omitempty"`
	ImageId                  string        `json:"imageId,omitempty"` // This bug must be fixed in CloudMC API
	TemplateName             string        `json:"templateName,omitempty"`
	IsPasswordEnabled        bool          `json:"isPasswordEnabled,omitempty"`
	IsSSHKeyEnabled          bool          `json:"isSshKeyEnabled,omitempty"`
	Username                 string        `json:"username,omitempty"`
	Password                 string        `json:"password,omitempty"`
	SSHKeyName               string        `json:"sshKeyName,omitempty"`
	Hypervisor               string        `json:"hypervisor,omitempty"`
	ComputeOfferingId        string        `json:"computeOfferingId,omitempty"`
	ComputeOfferingName      string        `json:"computeOfferingName,omitempty"`
	NewComputeOfferingId     string        `json:"newComputeOfferingId,omitempty"`
	CpuCount                 int           `json:"bareMetalCpuCount,omitempty"`
	MemoryInMB               int           `json:"memoryInMB,omitempty"`
	ZoneId                   string        `json:"zoneId,omitempty"`
	ZoneName                 string        `json:"zoneName,omitempty"`
	ProjectId                string        `json:"projectId,omitempty"`
	NetworkId                string        `json:"networkId,omitempty"`
	NetworkName              string        `json:"networkName,omitempty"`
	VpcId                    string        `json:"vpcId,omitempty"`
	VpcName                  string        `json:"vpcName,omitempty"`
	MacAddress               string        `json:"macAddress,omitempty"`
	UserData                 string        `json:"userData,omitempty"`
	RecoveryPoint            RecoveryPoint `json:"recoveryPoint,omitempty"`
	IpAddress                string        `json:"ipAddress,omitempty"`
	IpAddressId              string        `json:"ipAddressId,omitempty"`
	PublicIps                []PublicIp    `json:"publicIPs,omitempty"`
	PublicKey                string        `json:"publicKey,omitempty"`
	AdditionalDiskOfferingId string        `json:"diskOfferingId,omitempty"`
	AdditionalDiskSizeInGb   string        `json:"additionalDiskSizeInGb,omitempty"`
	AdditionalDiskIops       string        `json:"additionalDiskIops,omitempty"`
	VolumeIdToAttach         string        `json:"volumeIdToAttach,omitempty"`
	PortsToForward           []string      `json:"portsToForward,omitempty"`
	//RootVolumeSizeInGb       int           `json:"rootVolumeSizeInGb,omitempty"`
	DedicatedGroupId string   `json:"dedicatedGroupId,omitempty"`
	AffinityGroupIds []string `json:"affinityGroupIds,omitempty"`
}

func (baremetal *Baremetal) IsRunning() bool {
	return strings.EqualFold(baremetal.State, BAREMETAL_STATE_RUNNING)
}

func (baremetal *Baremetal) IsStopped() bool {
	return strings.EqualFold(baremetal.State, BAREMETAL_STATE_STOPPED)
}

type BaremetalService interface {
	Get(id string) (*Baremetal, error)
	List() ([]Baremetal, error)
	ListWithOptions(options map[string]string) ([]Baremetal, error)
	Create(Baremetal) (*Baremetal, error)
	Destroy(id string) (bool, error)
	Recover(id string) (bool, error)
	Exists(id string) (bool, error)
	Start(id string) (bool, error)
	Stop(id string) (bool, error)
	AssociateSSHKey(id string, sshKeyName string) (bool, error)
	Reboot(id string) (bool, error)
}

type BaremetalApi struct {
	entityService services.EntityService
}

func NewBaremetalService(apiClient api.ApiClient, serviceCode string, environmentName string) BaremetalService {
	return &BaremetalApi{
		entityService: services.NewEntityService(apiClient, serviceCode, environmentName, BAREMETAL_ENTITY_TYPE),
	}
}

func parseBaremetal(data []byte) *Baremetal {
	baremetal := Baremetal{}
	json.Unmarshal(data, &baremetal)
	return &baremetal
}

func parseBaremetalList(data []byte) []Baremetal {
	baremetals := []Baremetal{}
	json.Unmarshal(data, &baremetals)
	return baremetals
}

// Get baremetal with the specified id for the current environment
func (BaremetalApi *BaremetalApi) Get(id string) (*Baremetal, error) {
	data, err := BaremetalApi.entityService.Get(id, map[string]string{})
	if err != nil {
		return nil, err
	}
	return parseBaremetal(data), nil
}

// List all baremetals for the current environment
func (BaremetalApi *BaremetalApi) List() ([]Baremetal, error) {
	return BaremetalApi.ListWithOptions(map[string]string{})
}

// List all baremetals for the current environment. Can use options to do sorting and paging.
func (BaremetalApi *BaremetalApi) ListWithOptions(options map[string]string) ([]Baremetal, error) {
	data, err := BaremetalApi.entityService.List(options)
	if err != nil {
		return nil, err
	}
	return parseBaremetalList(data), nil
}

// Create a baremetal in the current environment
func (BaremetalApi *BaremetalApi) Create(baremetal Baremetal) (*Baremetal, error) {
	send, merr := json.Marshal(baremetal)
	if merr != nil {
		return nil, merr
	}
	optionsCopy := map[string]string{}
	optionsCopy["operation"] = "acquireBareMetal"

	body, err := BaremetalApi.entityService.Create(send, optionsCopy)
	if err != nil {
		return nil, err
	}
	return parseBaremetal(body), nil
}

// Destroy a baremetal with specified id in the current environment
func (BaremetalApi *BaremetalApi) Destroy(id string) (bool, error) {
	_, err := BaremetalApi.entityService.Execute(id, BAREMETAL_PURGE_OPERATION, []byte{}, map[string]string{})
	return err == nil, err
}

// Recover a destroyed baremetal with the specified id in the current environment
// Note: Cannot recover baremetals that have been purged
func (BaremetalApi *BaremetalApi) Recover(id string) (bool, error) {
	_, err := BaremetalApi.entityService.Execute(id, BAREMETAL_RECOVER_OPERATION, []byte{}, map[string]string{})
	return err == nil, err
}

// Check if baremetal with specified id exists in the current environment
func (BaremetalApi *BaremetalApi) Exists(id string) (bool, error) {
	_, err := BaremetalApi.Get(id)
	if err != nil {
		if hciError, ok := err.(api.HciErrorResponse); ok && hciError.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Start a stopped baremetal with specified id exists in the current environment
func (BaremetalApi *BaremetalApi) Start(id string) (bool, error) {
	_, err := BaremetalApi.entityService.Execute(id, BAREMETAL_START_OPERATION, []byte{}, map[string]string{})
	return err == nil, err
}

// Stop a running baremetal with specified id exists in the current environment
func (BaremetalApi *BaremetalApi) Stop(id string) (bool, error) {
	_, err := BaremetalApi.entityService.Execute(id, BAREMETAL_STOP_OPERATION, []byte{}, map[string]string{})
	return err == nil, err
}

// Associate an SSH key to the baremetal with the specified id exists in the current environment
// Note: This will reboot your baremetal if running
func (BaremetalApi *BaremetalApi) AssociateSSHKey(id string, sshKeyName string) (bool, error) {
	send, merr := json.Marshal(Baremetal{
		SSHKeyName: sshKeyName,
	})
	if merr != nil {
		return false, merr
	}
	_, err := BaremetalApi.entityService.Execute(id, BAREMETAL_ASSOCIATE_SSH_KEY_OPERATION, send, map[string]string{})
	return err == nil, err
}

// Reboot a running baremetal with specified id exists in the current environment
func (BaremetalApi *BaremetalApi) Reboot(id string) (bool, error) {
	_, err := BaremetalApi.entityService.Execute(id, BAREMETAL_REBOOT_OPERATION, []byte{}, map[string]string{})
	return err == nil, err
}
