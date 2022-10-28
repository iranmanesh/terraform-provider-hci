package hci

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	hci "github.com/hypertec-cloud/go-hci"
)

func TestAccNetworkCreate(t *testing.T) {
	t.Parallel()

	networkName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkCreateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkCreate(environmentID, vpcID, networkName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkCreateExists("hci_network.foobar"),
				),
			},
		},
	})
}

func testAccNetworkCreate(environment, vpc, name string) string {
	return fmt.Sprintf(`
resource "hci_network" "foobar" {
	environment_id   = "%s"
	vpc_id           = "%s"
	name             = "%s"
	description      = "This is a %s network"
	network_offering = "DefaultIsolatedNetworkOfferingForVpcNetworks"
	network_acl      = "default_allow"
}`, environment, vpc, name, name)
}

func testAccCheckNetworkCreateExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if rs.Primary.Attributes["environment_id"] == "" {
			return fmt.Errorf("Environment ID is missing")
		}

		client := testAccProvider.Meta().(*hci.HciClient)
		resources, err := getResourcesForEnvironmentID(client, rs.Primary.Attributes["environment_id"])
		if err != nil {
			return err
		}

		found, err := resources.Networks.Get(rs.Primary.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Network not found")
		}

		return nil
	}
}

func testAccCheckNetworkCreateDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hci.HciClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "hci_network" {
			if rs.Primary.Attributes["environment_id"] == "" {
				return fmt.Errorf("Environment ID is missing")
			}

			resources, err := getResourcesForEnvironmentID(client, rs.Primary.Attributes["environment_id"])
			if err != nil {
				return err
			}

			_, err = resources.Networks.Get(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Network still exists")
			}
		}
	}

	return nil
}
