package hci

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	hci "github.com/hypertec-cloud/go-hci"
)

func TestAccRemoteAccessVPNEnable(t *testing.T) {
	/*
		test is run in series since it uses a vpn that changes
		in another test
	*/

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRemoteAccessVPNEnableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRemoteAccessVPNEnable(environmentID, vpcID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRemoteAccessVPNEnableExists("hci_vpn.foobar"),
				),
			},
		},
	})
}

func testAccRemoteAccessVPNEnable(environment, vpc string) string {
	return fmt.Sprintf(`
resource "hci_vpn" "foobar" {
	environment_id = "%s"
	vpc_id         = "%s"
}`, environment, vpc)
}

func testAccCheckRemoteAccessVPNEnableExists(n string) resource.TestCheckFunc {
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

		found, err := resources.RemoteAccessVpn.Get(rs.Primary.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Remote Access VPN not found")
		}

		return nil
	}
}

func testAccCheckRemoteAccessVPNEnableDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hci.HciClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "hci_vpn" {
			if rs.Primary.Attributes["environment_id"] == "" {
				return fmt.Errorf("Environment ID is missing")
			}

			resources, err := getResourcesForEnvironmentID(client, rs.Primary.Attributes["environment_id"])
			if err != nil {
				return err
			}

			found, er := resources.RemoteAccessVpn.Get(rs.Primary.ID)
			if er == nil && found.State != DISABLED {
				return fmt.Errorf("Remote Access VPN still exists")
			}
		}
	}

	return nil
}
