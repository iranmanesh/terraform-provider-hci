package hci

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hci "github.com/hypertec-cloud/go-hci"
	"github.com/hypertec-cloud/go-hci/configuration"
)

// role names and fields
const (
	// role name
	EnvironmentAdminRole = "Environment admin"
	UserRole             = "User"
	ReadOnlyRole         = "Read-only"

	//fields
	OrganizationCode  = "organization_code"
	ServiceCode       = "service_code"
	Name              = "name"
	Description       = "description"
	AdminRoleUsers    = "admin_role"
	UserRoleUsers     = "user_role"
	ReadOnlyRoleUsers = "read_only_role"
)

func resourceHciEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceHciEnvironmentCreate,
		Read:   resourceHciEnvironmentRead,
		Update: resourceHciEnvironmentUpdate,
		Delete: resourceHciEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			OrganizationCode: {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "Organization's entry point, i.e. <entry_point>.hypertec.cloud",
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			ServiceCode: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A hypertec service code",
			},
			Name: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of environment to be created. Must be lower case, contain alphanumeric charaters, underscores or dashes",
			},
			Description: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description for the environment",
			},
			AdminRoleUsers: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of users that will be given Environment Admin role",
			},
			UserRoleUsers: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of users that will be given User role",
			},
			ReadOnlyRoleUsers: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of users that will be given Read-only role",
			},
		},
	}
}

func resourceHciEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	hciClient := meta.(*hci.HciClient)
	environment, err := hciClient.Environments.Get(d.Id())
	if err != nil {
		return handleNotFoundError("Environment", false, err, d)
	}

	adminRoleUsers, userRoleUsers, readOnlyRoleUsers := getUsersFromRoles(environment)
	adminRole, _ := d.GetOk(AdminRoleUsers)
	userRole, _ := d.GetOk(UserRoleUsers)
	readOnlyRole, _ := d.GetOk(ReadOnlyRoleUsers)

	if err := d.Set(OrganizationCode, environment.Organization.EntryPoint); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set(ServiceCode, environment.ServiceConnection.ServiceCode); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set(Name, environment.Name); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set(Description, environment.Description); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set(AdminRoleUsers, getListOfUsersByIDOrUsername(adminRoleUsers, adminRole.(*schema.Set))); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set(UserRoleUsers, getListOfUsersByIDOrUsername(userRoleUsers, userRole.(*schema.Set))); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	if err := d.Set(ReadOnlyRoleUsers, getListOfUsersByIDOrUsername(readOnlyRoleUsers, readOnlyRole.(*schema.Set))); err != nil {
		return fmt.Errorf("Error reading Trigger: %s", err)
	}

	return nil
}

func resourceHciEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	hciClient := meta.(*hci.HciClient)

	environment, err := getEnvironmentFromConfig(hciClient, d)
	if err != nil {
		return fmt.Errorf("Error parsing environment %s: %s", environment.Name, err)
	}

	newEnvironment, newErr := hciClient.Environments.Create(*environment)
	if newErr != nil {
		return fmt.Errorf("Error creating the new environment %s: %s", environment.Name, newErr)
	}

	d.SetId(newEnvironment.Id)

	return resourceHciEnvironmentRead(d, meta)
}

func resourceHciEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	hciClient := meta.(*hci.HciClient)
	environment, err := getEnvironmentFromConfig(hciClient, d)
	if err != nil {
		return fmt.Errorf("Error parsing environment %s: %s", environment.Name, err)
	}
	_, uerr := hciClient.Environments.Update(d.Id(), *environment)
	if uerr != nil {
		return fmt.Errorf("Error updating environment %s: %s", environment.Name, uerr)
	}
	return resourceHciEnvironmentRead(d, meta)
}

func resourceHciEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	hciClient := meta.(*hci.HciClient)
	fmt.Printf("[INFO] Destroying environment: %s\n", d.Get(Name).(string))
	if _, err := hciClient.Environments.Delete(d.Id()); err != nil {
		return handleNotFoundError("Environment", true, err, d)
	}
	return nil
}

func getEnvironmentFromConfig(hciClient *hci.HciClient, d *schema.ResourceData) (*configuration.Environment, error) {
	environment := configuration.Environment{}
	environment.Name = d.Get(Name).(string)
	environment.Description = d.Get(Description).(string)

	organizationID, oerr := getOrganizationID(hciClient, d.Get(OrganizationCode).(string))
	if oerr != nil {
		return &environment, oerr
	}

	connectionID, cerr := getServiceConnectionID(hciClient, d.Get(ServiceCode).(string))
	if cerr != nil {
		return &environment, cerr
	}

	environment.Organization = configuration.Organization{Id: organizationID}
	environment.ServiceConnection = configuration.ServiceConnection{Id: connectionID}

	adminRole, adminRoleExists := d.GetOk(AdminRoleUsers)
	userRole, userRoleExists := d.GetOk(UserRoleUsers)
	readOnlyRole, readOnlyRoleExists := d.GetOk(ReadOnlyRoleUsers)

	if adminRoleExists || userRoleExists || readOnlyRoleExists {

		users, uerr := hciClient.Users.ListWithOptions(map[string]string{"organizationId": organizationID})
		if uerr != nil {
			return &environment, uerr
		}

		environment.Roles = []configuration.Role{}

		if adminRoleExists {
			role, err := mapUsersToRole(EnvironmentAdminRole, adminRole.(*schema.Set).List(), users)
			if err != nil {
				return &environment, err
			}
			environment.Roles = append(environment.Roles, role)
		}

		if userRoleExists {
			role, err := mapUsersToRole(UserRole, userRole.(*schema.Set).List(), users)
			if err != nil {
				return &environment, err
			}
			environment.Roles = append(environment.Roles, role)
		}

		if readOnlyRoleExists {
			role, err := mapUsersToRole(ReadOnlyRole, readOnlyRole.(*schema.Set).List(), users)
			if err != nil {
				return &environment, err
			}
			environment.Roles = append(environment.Roles, role)
		}
	}
	return &environment, nil
}

func getListOfUsersByIDOrUsername(roleUsers []configuration.User, usersWithIDOrName *schema.Set) *schema.Set {
	mappedList := []interface{}{}
	for _, user := range roleUsers {
		found := false
		for _, idOrUsername := range usersWithIDOrName.List() {
			if isID(idOrUsername.(string)) {
				if strings.EqualFold(user.Id, idOrUsername.(string)) {
					found = true
					mappedList = append(mappedList, user.Id)
					break
				}
			} else if strings.EqualFold(user.Username, idOrUsername.(string)) {
				found = true
				mappedList = append(mappedList, user.Username)
				break
			}
		}
		if !found {
			mappedList = append(mappedList, user.Username)
		}
	}
	return schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), mappedList)
}

func getUsersFromRoles(environment *configuration.Environment) (adminRoleUsers []configuration.User, userRoleUsers []configuration.User, readOnlyRoleUsers []configuration.User) {
	for _, envRole := range environment.Roles {
		switch {
		case strings.EqualFold(envRole.Name, EnvironmentAdminRole):
			adminRoleUsers = append(adminRoleUsers, envRole.Users...)
		case strings.EqualFold(envRole.Name, UserRole):
			userRoleUsers = append(userRoleUsers, envRole.Users...)
		case strings.EqualFold(envRole.Name, ReadOnlyRole):
			readOnlyRoleUsers = append(readOnlyRoleUsers, envRole.Users...)
		}
	}
	return
}

func mapUsersToRole(roleName string, userList []interface{}, users []configuration.User) (configuration.Role, error) {
	role := configuration.Role{
		Name:  roleName,
		Users: []configuration.User{},
	}

	for _, userToFind := range userList {
		if isID(userToFind.(string)) {
			role.Users = append(role.Users, configuration.User{Id: userToFind.(string)})
			continue
		}
		found := false
		for _, user := range users {
			if strings.EqualFold(user.Username, userToFind.(string)) {
				found = true
				role.Users = append(role.Users, configuration.User{Id: user.Id})
				break
			}
		}
		if !found {
			return configuration.Role{}, fmt.Errorf("User %s was not found", userToFind)
		}
	}
	return role, nil
}

func getServiceConnectionID(hciClient *hci.HciClient, serviceCode string) (id string, err error) {
	if isID(serviceCode) {
		return serviceCode, nil
	}
	connections, cerr := hciClient.ServiceConnections.List()
	if cerr != nil {
		return "", cerr
	}
	for _, connection := range connections {
		if strings.EqualFold(connection.ServiceCode, serviceCode) {
			log.Printf("Found service connection : %+v", connection)
			return connection.Id, nil
		}
	}
	return "", nil
}

func getOrganizationID(hciClient *hci.HciClient, entryPoint string) (id string, err error) {
	if isID(entryPoint) {
		return entryPoint, nil
	}
	orgs, err := hciClient.Organizations.List()
	if err != nil {
		return "", err
	}
	for _, org := range orgs {
		if strings.EqualFold(org.EntryPoint, entryPoint) {
			log.Printf("Found organization: %+v", org)
			return org.Id, nil
		}
	}
	return "", fmt.Errorf("Organization with entry point %s not found", entryPoint)
}
