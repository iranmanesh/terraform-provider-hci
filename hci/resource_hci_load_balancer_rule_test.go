package hci

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	hci "github.com/hypertec-cloud/go-hci"
)

func TestAccLoadBalancerRuleCreate(t *testing.T) {
	t.Parallel()

	instanceName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLoadBalancerRuleCreateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerRuleCreate(environmentID, vpcID, networkID, instanceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerRuleCreateExists("hci_load_balancer_rule.foobar"),
				),
			},
		},
	})
}

func testAccLoadBalancerRuleCreate(environment, vpc, network, name string) string {
	return fmt.Sprintf(`
resource "hci_instance" "foobar" {
	environment_id   = "%s"
	network_id       = "%s"
	name             = "%s"
	template         = "Ubuntu 20.04.2"
	compute_offering = "Standard"
	cpu_count        = 1
	memory_in_mb     = 1024
}
resource "hci_public_ip" "foobar" {
	environment_id = "%s"
	vpc_id         = "%s"
}
resource "hci_load_balancer_rule" "foobar" {
	environment_id = "%s"
    network_id     = "%s"
    name           = "%s"
	public_ip_id   = "${hci_public_ip.foobar.id}"
    protocol       = "tcp"
    algorithm      = "leastconn"
    public_port    = 80
    private_port   = 80
    instance_ids   = ["${hci_instance.foobar.id}"]
}`, environment, network, name, environment, vpc, environment, network, name)
}

func testAccCheckLoadBalancerRuleCreateExists(n string) resource.TestCheckFunc {
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

		found, err := resources.LoadBalancerRules.Get(rs.Primary.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Load Balancer Rule not found")
		}

		return nil
	}
}

func testAccCheckLoadBalancerRuleCreateDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hci.HciClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "hci_load_balancer_rule" {
			if rs.Primary.Attributes["environment_id"] == "" {
				return fmt.Errorf("Environment ID is missing")
			}

			resources, err := getResourcesForEnvironmentID(client, rs.Primary.Attributes["environment_id"])
			if err != nil {
				return err
			}

			_, err = resources.LoadBalancerRules.Get(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("Load Balancer Rule still exists")
			}
		}
	}

	return nil
}
