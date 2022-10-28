package hci

import (
	"github.com/hypertec-cloud/go-hci/api"
	"github.com/hypertec-cloud/go-hci/configuration"
	"github.com/hypertec-cloud/go-hci/services"
	"github.com/hypertec-cloud/go-hci/services/hci"
)

const (
	DEFAULT_API_URL = "https://hypertec.cloud/api/v1/"
)

type HciClient struct {
	apiClient          api.ApiClient
	Tasks              services.TaskService
	Environments       configuration.EnvironmentService
	Users              configuration.UserService
	ServiceConnections configuration.ServiceConnectionService
	Organizations      configuration.OrganizationService
}

// Create a HciClient with the default URL
func NewHciClient(apiKey string) *HciClient {
	return NewHciClientWithURL(DEFAULT_API_URL, apiKey)
}

// Create a HciClient with a custom URL
func NewHciClientWithURL(apiURL string, apiKey string) *HciClient {
	apiClient := api.NewApiClient(apiURL, apiKey)
	return NewHciClientWithApiClient(apiClient)
}

// Create a HciClient with a custom URL that accepts insecure connections
func NewInsecureHciClientWithURL(apiURL string, apiKey string) *HciClient {
	apiClient := api.NewInsecureApiClient(apiURL, apiKey)
	return NewHciClientWithApiClient(apiClient)
}

func NewHciClientWithApiClient(apiClient api.ApiClient) *HciClient {
	hciClient := HciClient{
		apiClient:          apiClient,
		Tasks:              services.NewTaskService(apiClient),
		Environments:       configuration.NewEnvironmentService(apiClient),
		Users:              configuration.NewUserService(apiClient),
		ServiceConnections: configuration.NewServiceConnectionService(apiClient),
		Organizations:      configuration.NewOrganizationService(apiClient),
	}
	return &hciClient
}

// Get the Resources for a specific serviceCode and environmentName
// For now it assumes that the serviceCode belongs to a hci service type
func (c HciClient) GetResources(serviceCode string, environmentName string) (services.ServiceResources, error) {
	//TODO: change to check service type of service code
	return hci.NewResources(c.apiClient, serviceCode, environmentName), nil
}

// Get the API url used to do he calls
func (c HciClient) GetApiURL() string {
	return c.apiClient.GetApiURL()
}

// Get the API key used in the calls
func (c HciClient) GetApiKey() string {
	return c.apiClient.GetApiKey()
}

// Get the API Client used by all the services
func (c HciClient) GetApiClient() api.ApiClient {
	return c.apiClient
}
