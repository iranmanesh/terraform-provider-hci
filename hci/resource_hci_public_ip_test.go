package hci

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	hci "github.com/hypertec-cloud/go-hci"
)

func TestAccPublicIPCreate(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPublicIPCreateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicIPCreate(environmentID, vpcID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicIPCreateExists("hci_public_ip.foobar"),
				),
			},
		},
	})
}

func testAccPublicIPCreate(environment, vpc string) string {
	return fmt.Sprintf(`
resource "hci_public_ip" "foobar" {
	environment_id = "%s"
	vpc_id         = "%s"
}`, environment, vpc)
}

func testAccCheckPublicIPCreateExists(n string) resource.TestCheckFunc {
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

		found, err := resources.PublicIps.Get(rs.Primary.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Public IP not found")
		}

		return nil
	}
}

func testAccCheckPublicIPCreateDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hci.HciClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "hci_public_ip" {
			if rs.Primary.Attributes["environment_id"] == "" {
				return fmt.Errorf("Environment ID is missing")
			}

			resources, err := getResourcesForEnvironmentID(client, rs.Primary.Attributes["environment_id"])
			if err != nil {
				return err
			}

			_, err = resources.PublicIps.Get(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Public IP still exists")
			}
		}
	}

	return nil
}
