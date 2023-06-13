package hci

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	hci "github.com/hypertec-cloud/go-hci"
)

const hciBaremetal = "hci_baremetal"

func TestAccBaremetalCreateBasic(t *testing.T) {
	t.Parallel()

	baremetalName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBaremetalCreateBasicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBaremetalCreateBasic(environmentID, networkID, baremetalName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalCreateBasicExists("hci_baremetal.foobar"),
				),
			},
		},
	})
}

func TestAccBaremetalCreateDataDrive(t *testing.T) {
	t.Parallel()

	networkID := "e8360aac-cb3c-44cd-abfa-80701290e862"
	baremetalName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBaremetalCreateDataDriveDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBaremetalCreateDataDrive(environmentID, networkID, baremetalName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBaremetalCreateDataDriveExists("hci_baremetal.foobar"),
				),
			},
		},
	})
}

func testAccBaremetalCreateBasic(environment, network, name string) string {
	return fmt.Sprintf(`
resource %s "foobar" {
	environment_id   = "%s"
	network_id       = "%s"
	name             = "%s"
	template         = "CentOS 7.9 (bare-metal)"
	compute_offering = "metal.min.32.768.g02b"
}`, hciBaremetal, environment, network, name)
}

func testAccBaremetalCreateDataDrive(environment, network, name string) string {
	return fmt.Sprintf(`
resource %s "foobar" {
	environment_id   = "%s"
	network_id       = "%s"
	name             = "%s"
	template         = "CentOS 7.9 (bare-metal)"
	compute_offering = "metal.min.32.768.g02b"
}`, hciBaremetal, environment, network, name)
}

func testAccCheckBaremetalCreateBasicExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
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

		found, err := resources.Baremetals.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID || found.Name != rs.Primary.Attributes["name"] {
			return fmt.Errorf("Baremetal not found")
		}

		return nil
	}
}

func testAccCheckBaremetalCreateDataDriveExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("cannot find %s in state", name)
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

		found, err := resources.Baremetals.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID || found.Name != rs.Primary.Attributes["name"] {
			return fmt.Errorf("Baremetal not found")
		}

		return nil
	}
}

func testAccCheckBaremetalCreateBasicDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hci.HciClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == hciBaremetal {
			if rs.Primary.Attributes["environment_id"] == "" {
				return fmt.Errorf("Environment ID is missing")
			}

			resources, err := getResourcesForEnvironmentID(client, rs.Primary.Attributes["environment_id"])
			if err != nil {
				return err
			}

			_, err = resources.Baremetals.Get(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Baremetal still exists")
			}
		}
	}

	return nil
}

func testAccCheckBaremetalCreateDataDriveDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hci.HciClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == hciBaremetal {
			if rs.Primary.Attributes["environment_id"] == "" {
				return fmt.Errorf("Environment ID is missing")
			}

			resources, err := getResourcesForEnvironmentID(client, rs.Primary.Attributes["environment_id"])
			if err != nil {
				return err
			}

			_, err = resources.Baremetals.Get(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Baremetal still exists")
			}
		}
	}

	return nil
}
